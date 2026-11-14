package phase

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	"log/slog"

	"github.com/nirantarasi/k3sctl/pkg/apis/k3sctl.k3sproject.io/v1beta1/cluster"
	"github.com/nirantarasi/k3sctl/pkg/node"
	"github.com/nirantarasi/k3sctl/pkg/retry"
)

type InitializeK3s struct {
	GenericPhase
	leader *cluster.Host
}

func (p *InitializeK3s) Title() string {
	return "Initialize K3s cluster"
}

func (p *InitializeK3s) Prepare() {
	p.leader = p.Config.Spec.K3sLeader()
}

func (p *InitializeK3s) ShouldRun() bool {
	if p.leader == nil {
		return false
	}
	return !p.leader.Metadata.K3sServiceRunning
}

func (p *InitializeK3s) Run(ctx context.Context) error {
	h := p.leader

	slog.Info(fmt.Sprintf("%s: installing K3s service (server, cluster-init)", h))
	if err := p.installK3sService(h); err != nil {
		return fmt.Errorf("%s: install K3s service: %w", h, err)
	}

	slog.Info(fmt.Sprintf("%s: starting K3s service", h))
	svc, err := h.Sudo().Service(h.K3sServiceName())
	if err != nil {
		return fmt.Errorf("%s: get service: %w", h, err)
	}
	go reconcileEtcdMemberPeerURL(ctx, h)
	if err := svc.Start(ctx); err != nil {
		return fmt.Errorf("%s: start service: %w", h, err)
	}

	slog.Info(fmt.Sprintf("%s: waiting for K3s to be ready", h))
	if err := p.waitForReady(ctx, h); err != nil {
		return fmt.Errorf("%s: wait for ready: %w", h, err)
	}

	tokenPath := h.Configurer.K3sTokenPath()
	output, err := h.Sudo().ExecOutput(fmt.Sprintf("cat %s", shellQuote(tokenPath)))
	if err != nil {
		return fmt.Errorf("%s: read node token: %w", h, err)
	}
	h.Metadata.NodeToken = strings.TrimSpace(output)
	h.Metadata.NeedsUpgrade = false
	slog.Info(fmt.Sprintf("%s: K3s cluster initialized", h))

	return nil
}

func (p *InitializeK3s) installK3sService(h *cluster.Host) error {
	installFlags := h.K3sInstallFlags()

	unitContent := k3sServerUnit(h.Configurer.K3sBinaryPath(), installFlags, h.Environment, hostExecStartPre(h)...)
	unitPath := h.K3sServiceUnitPath()

	if err := h.Sudo().FS().WriteFile(unitPath, []byte(unitContent), fs.FileMode(0644)); err != nil {
		return fmt.Errorf("write service unit: %w", err)
	}

	if err := h.Sudo().Exec("systemctl daemon-reload"); err != nil {
		return fmt.Errorf("daemon-reload: %w", err)
	}

	if err := h.Sudo().Exec(fmt.Sprintf("systemctl enable %s", h.K3sServiceName())); err != nil {
		return fmt.Errorf("enable service: %w", err)
	}

	return nil
}

func (p *InitializeK3s) waitForReady(ctx context.Context, h *cluster.Host) error {
	kc, err := kubeClientOrBuild(p.Config.Metadata.KubeClient, h)
	if err != nil {
		return fmt.Errorf("build kube client for readiness check: %w", err)
	}
	if p.Config.Metadata.KubeClient == nil {
		p.Config.Metadata.KubeClient = kc
	}
	return retry.Context(ctx, node.KubeNodeReadyFunc(kc, h.KubernetesNodeName()),
		retry.Timeout(5*time.Minute), retry.Interval(10*time.Second))
}

func hostExecStartPre(h *cluster.Host) []string {
	var pre []string
	if h.Localhost {
		ips := make(map[string]bool)
		if h.PrivateAddress != "" && h.PrivateAddress != "localhost" && h.PrivateAddress != "127.0.0.1" {
			ips[h.PrivateAddress] = true
		}
		if h.Metadata.K3sExistingConfig != "" {
			for line := range strings.SplitSeq(h.Metadata.K3sExistingConfig, "\n") {
				line = strings.TrimSpace(line)
				if after, ok := strings.CutPrefix(line, "node-ip:"); ok {
					val := strings.TrimSpace(after)
					val = strings.Trim(val, `"'`)
					if val != "" && val != "localhost" && val != "127.0.0.1" {
						ips[val] = true
					}
				}
			}
		}
		sortedIPs := make([]string, 0, len(ips))
		for ip := range ips {
			sortedIPs = append(sortedIPs, ip)
		}
		sort.Strings(sortedIPs)
		for _, ip := range sortedIPs {
			pre = append(pre, fmt.Sprintf("-/sbin/ip addr add %s/32 dev lo", ip))
		}
	}
	return pre
}

func k3sServerUnit(binaryPath string, flags []string, env map[string]string, extraExecStartPre ...string) string {
	var execStart strings.Builder
	execStart.WriteString(binaryPath + " server")
	for _, f := range flags {
		execStart.WriteString(" " + f)
	}
	var preCommands strings.Builder
	for _, pre := range extraExecStartPre {
		if strings.TrimSpace(pre) != "" {
			fmt.Fprintf(&preCommands, "ExecStartPre=%s\n", pre)
		}
	}
	return fmt.Sprintf(`[Unit]
Description=K3s - Lightweight Kubernetes
Documentation=https://k3s.io
After=network-online.target
Wants=network-online.target

[Service]
Type=notify
%s%sExecStartPre=-/sbin/modprobe br_netfilter
ExecStartPre=-/sbin/modprobe overlay
ExecStart=%s
KillMode=process
Delegate=yes
LimitNOFILE=1048576
LimitNPROC=infinity
LimitCORE=infinity
TasksMax=infinity
TimeoutStartSec=0
Restart=always
RestartSec=5s

[Install]
WantedBy=multi-user.target
`, systemdEnvironment(env), preCommands.String(), execStart.String())
}

func k3sAgentUnit(binaryPath string, flags []string, env map[string]string, extraExecStartPre ...string) string {
	var execStart strings.Builder
	execStart.WriteString(binaryPath + " agent")
	for _, f := range flags {
		execStart.WriteString(" " + f)
	}
	var preCommands strings.Builder
	for _, pre := range extraExecStartPre {
		if strings.TrimSpace(pre) != "" {
			fmt.Fprintf(&preCommands, "ExecStartPre=%s\n", pre)
		}
	}
	return fmt.Sprintf(`[Unit]
Description=K3s Agent - Lightweight Kubernetes
Documentation=https://k3s.io
After=network-online.target
Wants=network-online.target

[Service]
Type=exec
%s%sExecStart=%s
KillMode=process
Delegate=yes
LimitNOFILE=1048576
LimitNPROC=infinity
LimitCORE=infinity
TasksMax=infinity
TimeoutStartSec=0
Restart=always
RestartSec=5s

[Install]
WantedBy=multi-user.target
`, systemdEnvironment(env), preCommands.String(), execStart.String())
}

// systemdEnvironment renders the cluster's environment variables as
// Environment= directives. Keys are sorted so the generated unit is stable.
func systemdEnvironment(env map[string]string) string {
	if len(env) == 0 {
		return ""
	}
	keys := make([]string, 0, len(env))
	for key := range env {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, key := range keys {
		fmt.Fprintf(&b, "Environment=%q\n", key+"="+env[key])
	}
	return b.String()
}
