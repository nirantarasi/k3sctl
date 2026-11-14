package phase

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/nirantarasi/k3sctl/pkg/apis/k3sctl.k3sproject.io/v1beta1/cluster"
)

type Lock struct {
	GenericPhase
}

func (p *Lock) Title() string {
	return "Acquire exclusive host lock"
}

func (p *Lock) Run(ctx context.Context) error {
	return p.parallelDo(ctx, p.Config.Spec.Hosts, func(ctx context.Context, h *cluster.Host) error {
		lockFile := h.Configurer.K3sctlLockFilePath()
		if h.Sudo().FS().FileExist(lockFile) {
			stat, err := h.Sudo().FS().Stat(lockFile)
			if err == nil {
				// If lock file is older than 30 seconds, treat it as stale from a crashed/restarted run and auto-remove it.
				if time.Since(stat.ModTime()) > 30*time.Second {
					slog.Info(fmt.Sprintf("%s: removing stale lock file (older than 30s): %s", h, lockFile))
					_ = h.Sudo().FS().Remove(lockFile)
				} else {
					content, err := h.Sudo().ExecOutput(fmt.Sprintf("cat %s", shellQuote(lockFile)))
					if err == nil && content != "" {
						return fmt.Errorf("%s: host is locked by another k3sctl operation (lock: %s)", h, content)
					}
				}
			}
		}

		if err := h.Sudo().Exec(fmt.Sprintf("mkdir -p %s && echo k3sctl > %s",
			shellQuote(h.Configurer.K3sDataDirDefaultPath()),
			shellQuote(lockFile),
		)); err != nil {
			return fmt.Errorf("%s: acquire lock: %w", h, err)
		}

		slog.Debug(fmt.Sprintf("%s: acquired lock", h))
		return nil
	})
}
