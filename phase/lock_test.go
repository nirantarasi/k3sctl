package phase

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	rig "github.com/k0sproject/rig/v2"
	"github.com/k0sproject/rig/v2/cmd"
	"github.com/k0sproject/rig/v2/rigtest"
	"github.com/nirantarasi/k3sctl/configurer/mocks"
	"github.com/nirantarasi/k3sctl/pkg/apis/k3sctl.k3sproject.io/v1beta1"
	"github.com/nirantarasi/k3sctl/pkg/apis/k3sctl.k3sproject.io/v1beta1/cluster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The lock and unlock phases talk to the host through rig: existence with
// `test -f`, the age of a lock through `stat`, the holder through `cat`, the
// lock itself through `mkdir -p … && echo k3sctl > …`, and release through
// `rm -f`. These tests drive the real phase code with a mock connection that
// answers those commands, so no host is needed.

const (
	testLockFile = "/var/lock/k3sctl.lock"
	testDataDir  = "/var/lib/rancher/k3s"
	// The line rig's PosixFS parses: mode, size, mtime as seconds.nanoseconds, name.
	statLine = "0x81a4 0 %d.%09d //%s//"
)

// hostWithLock builds a host whose lock file either exists (with the given
// mtime and holder) or does not, and records the commands the phase runs.
func hostWithLock(t *testing.T, exists bool, modTime time.Time, holder string) (*cluster.Host, *[]string) {
	t.Helper()
	conn := rigtest.NewMockConnection()
	var ran []string
	conn.AddCommand(rigtest.Contains("test -f"), func(a *rigtest.A) error {
		ran = append(ran, a.Command)
		if !exists {
			return fmt.Errorf("exit status 1") // rig's FileExist is `test -f …` returning an error
		}
		return nil
	})
	conn.AddCommand(rigtest.Contains("stat --help"), func(a *rigtest.A) error {
		ran = append(ran, a.Command)
		fmt.Fprint(a.Stdout, "--format=") // GNU stat, which rig picks by this help text
		return nil
	})
	conn.AddCommand(rigtest.Contains("stat -c"), func(a *rigtest.A) error {
		ran = append(ran, a.Command)
		fmt.Fprintf(a.Stdout, statLine, modTime.Unix(), modTime.Nanosecond(), testLockFile)
		return nil
	})
	conn.AddCommand(rigtest.Contains("cat "), func(a *rigtest.A) error {
		ran = append(ran, a.Command)
		fmt.Fprintf(a.Stdout, "%s\n", holder)
		return nil
	})
	conn.AddCommand(rigtest.Contains("mkdir -p"), func(a *rigtest.A) error {
		ran = append(ran, a.Command)
		return nil
	})
	conn.AddCommand(rigtest.Contains("rm -f"), func(a *rigtest.A) error {
		ran = append(ran, a.Command)
		return nil
	})

	client, err := rig.NewClient(rig.WithConnection(conn), rig.WithRunner(cmd.NewExecutor(conn)))
	require.NoError(t, err)

	cfg := &mocks.FakeConfigurer{}
	cfg.K3sctlLockFilePathReturns(testLockFile)
	cfg.K3sDataDirDefaultPathReturns(testDataDir)

	return &cluster.Host{Client: client, Configurer: cfg}, &ran
}

func newLock(host *cluster.Host) *Lock {
	p := &Lock{}
	p.Config = &v1beta1.Cluster{Spec: &cluster.Spec{Hosts: cluster.Hosts{host}}}
	return p
}

func newUnlock(host *cluster.Host) *Unlock {
	p := &Unlock{}
	p.Config = &v1beta1.Cluster{Spec: &cluster.Spec{Hosts: cluster.Hosts{host}}}
	return p
}

func TestLockTitles(t *testing.T) {
	assert.Equal(t, "Acquire exclusive host lock", newLock(nil).Title())
	assert.Equal(t, "Release exclusive host lock", newUnlock(nil).Title())
}

func TestLockAcquiresWhenThereIsNoLockFile(t *testing.T) {
	host, ran := hostWithLock(t, false, time.Time{}, "")
	require.NoError(t, newLock(host).Run(context.Background()))

	// The phase shell-quotes both paths, so match on the paths themselves.
	var create string
	for _, c := range *ran {
		if strings.Contains(c, "mkdir -p") {
			create = c
		}
	}
	require.Contains(t, *ran, create, "the lock is taken with a command")
	assert.Contains(t, create, "mkdir -p "+"'"+testDataDir+"'", "it creates the data directory: %q", create)
	assert.Contains(t, create, "echo k3sctl > "+"'"+testLockFile+"'", "and writes the lock file: %q", create)
	assert.NotContains(t, strings.Join(*ran, " "), "cat ", "there is nothing to read when the file is missing")
}

func TestLockRefusesAHeldLock(t *testing.T) {
	host, _ := hostWithLock(t, true, time.Now().Add(-2*time.Second), "k3sctl")
	err := newLock(host).Run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "host is locked by another k3sctl operation")
	assert.Contains(t, err.Error(), "k3sctl", "the error names the holder it read from the lock file")
}

// A lock left behind by a run that crashed is removed instead of blocking the
// next one, and the removal happens before the new lock is written.
func TestLockRemovesAStaleLock(t *testing.T) {
	host, ran := hostWithLock(t, true, time.Now().Add(-time.Minute), "k3sctl")
	require.NoError(t, newLock(host).Run(context.Background()))

	removeAt, createAt := -1, -1
	for i, c := range *ran {
		if strings.Contains(c, "rm -f") && removeAt < 0 {
			removeAt = i
		}
		if strings.Contains(c, "mkdir -p") && createAt < 0 {
			createAt = i
		}
	}
	require.GreaterOrEqual(t, removeAt, 0, "the stale lock is removed: %v", *ran)
	require.GreaterOrEqual(t, createAt, 0, "the new lock is written: %v", *ran)
	assert.Less(t, removeAt, createAt, "the stale lock goes before the new one is written")
}

// The age boundary is 30 seconds: 29 is still held, 31 is stale.
func TestLockStaleBoundary(t *testing.T) {
	held, _ := hostWithLock(t, true, time.Now().Add(-29*time.Second), "k3sctl")
	assert.Error(t, newLock(held).Run(context.Background()), "29 seconds old is still a live lock")

	stale, _ := hostWithLock(t, true, time.Now().Add(-31*time.Second), "k3sctl")
	assert.NoError(t, newLock(stale).Run(context.Background()), "31 seconds old is stale")
}

func TestUnlockRemovesTheLockFile(t *testing.T) {
	host, ran := hostWithLock(t, true, time.Now(), "k3sctl")
	require.NoError(t, newUnlock(host).Run(context.Background()))
	assert.Contains(t, strings.Join(*ran, " "), "rm -f "+testLockFile)
}

// Unlocking when nothing is locked is not an error: the phase runs on release
// paths that may follow a phase that failed before taking the lock.
func TestUnlockWithoutALockFileDoesNothing(t *testing.T) {
	host, ran := hostWithLock(t, false, time.Time{}, "")
	require.NoError(t, newUnlock(host).Run(context.Background()))
	assert.NotContains(t, strings.Join(*ran, " "), "rm -f")
}
