package phase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"log/slog"

	"github.com/nirantarasi/k3sctl/internal/k3spath"
	"github.com/nirantarasi/k3sctl/pkg/apis/k3sctl.k3sproject.io/v1beta1/cluster"
	"github.com/nirantarasi/k3sctl/pkg/etcd"
)

// etcdClientOrBuild returns the existing etcd client from cluster metadata,
// or builds a new one by reading TLS certificates from the given host.
func etcdClientOrBuild(ctx context.Context, existing etcd.Client, h *cluster.Host) (etcd.Client, error) {
	if existing != nil {
		return existing, nil
	}
	return buildEtcdClient(ctx, h)
}

func buildEtcdClient(ctx context.Context, h *cluster.Host) (etcd.Client, error) {
	slog.Info(fmt.Sprintf("%s: reading etcd TLS certificates", h))

	caCert, err := h.Sudo().ExecOutput(fmt.Sprintf("cat %s", shellQuote(k3spath.EtcdCACertPath)))
	if err != nil {
		return nil, fmt.Errorf("%s: read etcd CA cert: %w", h, err)
	}

	cert, err := h.Sudo().ExecOutput(fmt.Sprintf("cat %s", shellQuote(k3spath.EtcdClientCertPath)))
	if err != nil {
		return nil, fmt.Errorf("%s: read etcd client cert: %w", h, err)
	}

	key, err := h.Sudo().ExecOutput(fmt.Sprintf("cat %s", shellQuote(k3spath.EtcdClientKeyPath)))
	if err != nil {
		return nil, fmt.Errorf("%s: read etcd client key: %w", h, err)
	}

	endpoint := fmt.Sprintf("https://%s:2379", etcdEndpointAddr(h))
	endpoints := []string{endpoint}
	if h.Localhost && etcdEndpointAddr(h) != "127.0.0.1" {
		endpoints = append(endpoints, "https://127.0.0.1:2379")
	}
	ec, err := etcd.New(ctx, endpoints, []byte(caCert), []byte(cert), []byte(key))
	if err != nil {
		return nil, fmt.Errorf("%s: create etcd client: %w", h, err)
	}

	slog.Info(fmt.Sprintf("%s: etcd client connected to %s", h, endpoint))
	return ec, nil
}

func etcdEndpointAddr(h *cluster.Host) string {
	if h.PrivateAddress != "" {
		return h.PrivateAddress
	}
	if addr := h.Address(); addr != "" {
		return addr
	}
	return "127.0.0.1"
}

// reconcileEtcdMemberPeerURL ensures that the server host's etcd member has a peer URL
// matching its desired PrivateAddress. If the host's IP address changed (e.g. DHCP
// lease renewal or Wi-Fi network change), etcd's stored member will still have the old peer URL,
// which prevents K3s from joining. This function polls local etcd, updates the member's
// peer URL to the new address, and allows K3s to finish booting cleanly.
func reconcileEtcdMemberPeerURL(ctx context.Context, h *cluster.Host) {
	if !h.IsServer() || h.PrivateAddress == "" {
		return
	}

	caCertPath := k3spath.EtcdCACertPath
	if !h.Sudo().FS().FileExist(caCertPath) {
		return
	}

	caCert, err := h.Sudo().ExecOutput(fmt.Sprintf("cat %s", shellQuote(caCertPath)))
	if err != nil || strings.TrimSpace(caCert) == "" {
		return
	}
	cert, err := h.Sudo().ExecOutput(fmt.Sprintf("cat %s", shellQuote(k3spath.EtcdClientCertPath)))
	if err != nil || strings.TrimSpace(cert) == "" {
		return
	}
	key, err := h.Sudo().ExecOutput(fmt.Sprintf("cat %s", shellQuote(k3spath.EtcdClientKeyPath)))
	if err != nil || strings.TrimSpace(key) == "" {
		return
	}

	desiredPeerURL := fmt.Sprintf("https://%s:2380", h.PrivateAddress)
	endpoints := []string{"https://127.0.0.1:2379"}
	if h.PrivateAddress != "127.0.0.1" {
		endpoints = append(endpoints, fmt.Sprintf("https://%s:2379", h.PrivateAddress))
	}

	// Retry connecting to etcd for up to 60 seconds while K3s starts it up
	connectCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-connectCtx.Done():
			return
		case <-ticker.C:
			ec, err := etcd.New(connectCtx, endpoints, []byte(caCert), []byte(cert), []byte(key))
			if err != nil {
				continue
			}

			members, err := ec.MemberList(connectCtx)
			if err != nil {
				_ = ec.Close()
				continue
			}

			var targetMember *etcd.Member
			nodeName := h.KubernetesNodeName()
			hostname := h.Metadata.Hostname

			for _, m := range members {
				if (nodeName != "" && strings.HasPrefix(m.Name, nodeName)) ||
					(hostname != "" && strings.HasPrefix(m.Name, hostname)) ||
					len(members) == 1 {
					targetMember = &m
					break
				}
			}

			if targetMember == nil {
				_ = ec.Close()
				return
			}

			hasPeer := false
			for _, u := range targetMember.PeerURLs {
				if u == desiredPeerURL {
					hasPeer = true
					break
				}
			}

			if !hasPeer {
				slog.Info(fmt.Sprintf("%s: reconciling etcd member %s (%x) peer URL to %s", h, targetMember.Name, targetMember.ID, desiredPeerURL))
				if err := ec.MemberUpdate(connectCtx, targetMember.ID, []string{desiredPeerURL}); err != nil {
					slog.Warn(fmt.Sprintf("%s: failed to update etcd member peer URL: %v", h, err))
				} else {
					slog.Info(fmt.Sprintf("%s: successfully updated etcd member peer URL to %s", h, desiredPeerURL))
				}
			}

			_ = ec.Close()
			return
		}
	}
}
