package phase

import (
	"context"
	"fmt"

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
	ec, err := etcd.New(ctx, []string{endpoint}, []byte(caCert), []byte(cert), []byte(key))
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
