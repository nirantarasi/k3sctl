package phase

import (
	"strings"
	"testing"

	"github.com/nirantarasi/k3sctl/pkg/apis/k3sctl.k3sproject.io/v1beta1/cluster"
	"github.com/stretchr/testify/assert"
)

func TestSystemdEnvironment(t *testing.T) {
	assert.Empty(t, systemdEnvironment(nil))

	got := systemdEnvironment(map[string]string{"B": "2", "A": "1"})
	assert.Equal(t, "Environment=\"A=1\"\nEnvironment=\"B=2\"\n", got)
}

func TestK3sServerUnitRendersEnvironment(t *testing.T) {
	unit := k3sServerUnit("/usr/local/bin/k3s", []string{"--cluster-init"}, map[string]string{
		"ETCD_UNSUPPORTED_ARCH": "arm",
	})

	assert.True(t, strings.Contains(unit, "Environment=\"ETCD_UNSUPPORTED_ARCH=arm\"\n"))
	assert.True(t, strings.HasPrefix(unit, "[Unit]\n"))
	assert.True(t, strings.Contains(unit, "ExecStart=/usr/local/bin/k3s server --cluster-init\n"))
}

func TestK3sAgentUnitRendersEnvironment(t *testing.T) {
	unit := k3sAgentUnit("/usr/local/bin/k3s", nil, map[string]string{"FOO": "bar"})

	assert.True(t, strings.Contains(unit, "Environment=\"FOO=bar\"\n"))
	assert.True(t, strings.Contains(unit, "ExecStart=/usr/local/bin/k3s agent\n"))
}

func TestHostExecStartPre_Localhost(t *testing.T) {
	h := &cluster.Host{
		Localhost:      true,
		PrivateAddress: "10.52.64.248",
		Metadata: cluster.HostMetadata{
			K3sExistingConfig: "node-ip: \"172.25.223.248\"\ncluster-init: true\n",
		},
	}

	pre := hostExecStartPre(h)
	assert.Contains(t, pre, "-/sbin/ip addr add 10.52.64.248/32 dev lo")
	assert.Contains(t, pre, "-/sbin/ip addr add 172.25.223.248/32 dev lo")

	unit := k3sServerUnit("/usr/local/bin/k3s", nil, nil, pre...)
	assert.True(t, strings.Contains(unit, "ExecStartPre=-/sbin/ip addr add 10.52.64.248/32 dev lo\n"))
	assert.True(t, strings.Contains(unit, "ExecStartPre=-/sbin/ip addr add 172.25.223.248/32 dev lo\n"))
}

func TestHostExecStartPre_NonLocalhost(t *testing.T) {
	h := &cluster.Host{
		PrivateAddress: "10.0.0.5",
	}
	assert.Empty(t, hostExecStartPre(h))
}
