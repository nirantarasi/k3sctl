package cluster

import (
	"fmt"
	"os"
	"strings"

	"github.com/k0sproject/dig"
	"gopkg.in/yaml.v2"
)

func fileExistsLocal(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

type EtcdSnapshotConfig struct {
	Scheduled    bool             `yaml:"scheduled,omitempty"`
	ScheduleCron string           `yaml:"scheduleCron,omitempty"`
	Retention    int              `yaml:"retention,omitempty"`
	Dir          string           `yaml:"dir,omitempty"`
	Compress     bool             `yaml:"compress,omitempty"`
	S3           *EtcdS3Config    `yaml:"s3,omitempty"`
	Prune        *EtcdPruneConfig `yaml:"prune,omitempty"`
}

type EtcdPruneConfig struct {
	Enabled   bool   `yaml:"enabled,omitempty"`
	Retention int    `yaml:"retention,omitempty"`
	Name      string `yaml:"name,omitempty"`
}

type EtcdS3Config struct {
	Enabled          bool   `yaml:"enabled,omitempty"`
	Endpoint         string `yaml:"endpoint,omitempty"`
	EndpointCA       string `yaml:"endpointCa,omitempty"`
	SkipSSLVerify    bool   `yaml:"skipSslVerify,omitempty"`
	AccessKey        string `yaml:"accessKey,omitempty"`
	SecretKey        string `yaml:"secretKey,omitempty"`
	Bucket           string `yaml:"bucket,omitempty"`
	Region           string `yaml:"region,omitempty"`
	Folder           string `yaml:"folder,omitempty"`
	Insecure         bool   `yaml:"insecure,omitempty"`
	SessionToken     string `yaml:"sessionToken,omitempty"`
	BucketLookupType string `yaml:"bucketLookupType,omitempty"`
	Proxy            string `yaml:"proxy,omitempty"`
	ConfigSecret     string `yaml:"configSecret,omitempty"`
	Timeout          string `yaml:"timeout,omitempty"`
	Retention        int    `yaml:"retention,omitempty"`
}

type EtcdAdvancedConfig struct {
	ExposeMetrics             bool   `yaml:"exposeMetrics,omitempty"`
	DisableSnapshots          bool   `yaml:"disableSnapshots,omitempty"`
	SnapshotName              string `yaml:"snapshotName,omitempty"`
	SnapshotReconcileInterval string `yaml:"snapshotReconcileInterval,omitempty"`
}

type SecretsEncryptConfig struct {
	Enabled    bool `yaml:"enabled,omitempty"`
	RotateKeys bool `yaml:"rotateKeys,omitempty"`
	Reencrypt  bool `yaml:"reencrypt,omitempty"`
}

type CertRotateConfig struct {
	Enabled  bool `yaml:"enabled,omitempty"`
	RotateCA bool `yaml:"rotateCA,omitempty"`
}

type TokenConfig struct {
	AgentToken     string `yaml:"agentToken,omitempty"`
	ServerToken    string `yaml:"serverToken,omitempty"`
	TokenFile      string `yaml:"tokenFile,omitempty"`
	AgentTokenFile string `yaml:"agentTokenFile,omitempty"`
	Rotate         bool   `yaml:"rotate,omitempty"`
}

type PreflightConfig struct {
	Enabled      bool `yaml:"enabled" default:"true"`
	CheckCgroups bool `yaml:"checkCgroups" default:"true"`
	CheckModules bool `yaml:"checkModules" default:"true"`
	CheckPorts   bool `yaml:"checkPorts" default:"true"`
	CheckDisk    bool `yaml:"checkDisk" default:"true"`
	MinDiskGB    int  `yaml:"minDiskGB,omitempty"`
}

type AirgapConfig struct {
	Enabled     bool   `yaml:"enabled"`
	ImageBundle string `yaml:"imageBundle,omitempty"`
}

type RegistriesConfig struct {
	Mirrors map[string]RegistryMirror `yaml:"mirrors,omitempty"`
	Configs map[string]RegistryAuth   `yaml:"configs,omitempty"`
}

type RegistryMirror struct {
	Endpoint []string `yaml:"endpoint,omitempty"`
}

type RegistryAuth struct {
	Auth *RegistryAuthConfig `yaml:"auth,omitempty"`
	TLS  *RegistryTLSConfig  `yaml:"tls,omitempty"`
}

type RegistryAuthConfig struct {
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Token    string `yaml:"token,omitempty"`
}

type RegistryTLSConfig struct {
	CertFile           string `yaml:"certFile,omitempty"`
	KeyFile            string `yaml:"keyFile,omitempty"`
	CAFile             string `yaml:"caFile,omitempty"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify,omitempty"`
}

type HealthCheckConfig struct {
	Enabled bool   `yaml:"enabled,omitempty"`
	Timeout string `yaml:"timeout,omitempty"`
}

type NetworkConfig struct {
	ClusterCIDR        []string `yaml:"clusterCIDR,omitempty"`
	ServiceCIDR        []string `yaml:"serviceCIDR,omitempty"`
	ClusterDNS         []string `yaml:"clusterDNS,omitempty"`
	ClusterDomain      string   `yaml:"clusterDomain,omitempty"`
	NodePortRange      string   `yaml:"nodePortRange,omitempty"`
	FlannelBackend     string   `yaml:"flannelBackend,omitempty"`
	FlannelIPv6Masq    bool     `yaml:"flannelIPv6Masq,omitempty"`
	FlannelExternalIP  bool     `yaml:"flannelExternalIP,omitempty"`
	FlannelIface       string   `yaml:"flannelIface,omitempty"`
	FlannelConf        string   `yaml:"flannelConf,omitempty"`
	FlannelCNIConf     string   `yaml:"flannelCNIConf,omitempty"`
	EgressSelectorMode string   `yaml:"egressSelectorMode,omitempty"`
	ServiceLBNamespace string   `yaml:"serviceLBNamespace,omitempty"`
}

type DisableConfig struct {
	Components      []string `yaml:"components,omitempty"`
	Scheduler       bool     `yaml:"scheduler,omitempty"`
	CloudController bool     `yaml:"cloudController,omitempty"`
	KubeProxy       bool     `yaml:"kubeProxy,omitempty"`
	NetworkPolicy   bool     `yaml:"networkPolicy,omitempty"`
	HelmController  bool     `yaml:"helmController,omitempty"`
}

type TLSConfig struct {
	SANs        []string `yaml:"sans,omitempty"`
	SANSecurity bool     `yaml:"sanSecurity,omitempty"`
}

type KubeconfigConfig struct {
	Mode  string `yaml:"mode,omitempty"`
	Group string `yaml:"group,omitempty"`
	Name  string `yaml:"name,omitempty"`
	Path  string `yaml:"path,omitempty"`
}

type DatastoreConfig struct {
	Endpoint string `yaml:"endpoint,omitempty"`
	CAFile   string `yaml:"caFile,omitempty"`
	CertFile string `yaml:"certFile,omitempty"`
	KeyFile  string `yaml:"keyFile,omitempty"`
}

type ServerConfig struct {
	AdvertiseAddress  string `yaml:"advertiseAddress,omitempty"`
	AdvertisePort     int    `yaml:"advertisePort,omitempty"`
	BindAddress       string `yaml:"bindAddress,omitempty"`
	HTTPSPort         int    `yaml:"httpsPort,omitempty"`
	SupervisorMetrics bool   `yaml:"supervisorMetrics,omitempty"`
	EnablePprof       bool   `yaml:"enablePprof,omitempty"`
}

type SecurityConfig struct {
	SELinux               bool   `yaml:"selinux,omitempty"`
	ProtectKernelDefaults bool   `yaml:"protectKernelDefaults,omitempty"`
	EncryptionProvider    string `yaml:"encryptionProvider,omitempty"`
	Rootless              bool   `yaml:"rootless,omitempty"`
	PreferBundledBin      bool   `yaml:"preferBundledBin,omitempty"`
}

type ComponentArgsConfig struct {
	KubeAPIServer              []string `yaml:"kubeAPIServer,omitempty"`
	KubeControllerManager      []string `yaml:"kubeControllerManager,omitempty"`
	KubeScheduler              []string `yaml:"kubeScheduler,omitempty"`
	KubeCloudControllerManager []string `yaml:"kubeCloudControllerManager,omitempty"`
	Kubelet                    []string `yaml:"kubelet,omitempty"`
	KubeProxy                  []string `yaml:"kubeProxy,omitempty"`
	Etcd                       []string `yaml:"etcd,omitempty"`
	HelmController             []string `yaml:"helmController,omitempty"`
}

type RuntimeConfig struct {
	SystemDefaultRegistry    string `yaml:"systemDefaultRegistry,omitempty"`
	PauseImage               string `yaml:"pauseImage,omitempty"`
	Snapshotter              string `yaml:"snapshotter,omitempty"`
	DefaultRuntime           string `yaml:"defaultRuntime,omitempty"`
	ResolvConf               string `yaml:"resolvConf,omitempty"`
	ContainerRuntimeEndpoint string `yaml:"containerRuntimeEndpoint,omitempty"`
	ImageServiceEndpoint     string `yaml:"imageServiceEndpoint,omitempty"`
}

type EmbeddedRegistryConfig struct {
	Enabled                        bool `yaml:"enabled,omitempty"`
	DisableDefaultRegistryEndpoint bool `yaml:"disableDefaultRegistryEndpoint,omitempty"`
}

type ImageCredentialProviderConfig struct {
	BinDir string `yaml:"binDir,omitempty"`
	Config string `yaml:"config,omitempty"`
}

type LocalStorageConfig struct {
	Path string `yaml:"path,omitempty"`
}

type AgentConfig struct {
	LBServerPort int `yaml:"lbServerPort,omitempty"`
}

type K3s struct {
	Version                 string                         `yaml:"version,omitempty"`
	Config                  dig.Mapping                    `yaml:"config,omitempty"`
	ExtraArgs               Flags                          `yaml:"extraArgs,omitempty"`
	EtcdSnapshots           *EtcdSnapshotConfig            `yaml:"etcdSnapshots,omitempty"`
	SecretsEncrypt          *SecretsEncryptConfig          `yaml:"secretsEncrypt,omitempty"`
	CertRotate              *CertRotateConfig              `yaml:"certRotate,omitempty"`
	Token                   *TokenConfig                   `yaml:"token,omitempty"`
	Preflight               *PreflightConfig               `yaml:"preflight,omitempty"`
	Airgap                  *AirgapConfig                  `yaml:"airgap,omitempty"`
	Registries              *RegistriesConfig              `yaml:"registries,omitempty"`
	HealthCheck             *HealthCheckConfig             `yaml:"healthCheck,omitempty"`
	Network                 *NetworkConfig                 `yaml:"network,omitempty"`
	Disable                 *DisableConfig                 `yaml:"disable,omitempty"`
	TLS                     *TLSConfig                     `yaml:"tls,omitempty"`
	Kubeconfig              *KubeconfigConfig              `yaml:"kubeconfig,omitempty"`
	Datastore               *DatastoreConfig               `yaml:"datastore,omitempty"`
	Server                  *ServerConfig                  `yaml:"server,omitempty"`
	Security                *SecurityConfig                `yaml:"security,omitempty"`
	ComponentArgs           *ComponentArgsConfig           `yaml:"componentArgs,omitempty"`
	Runtime                 *RuntimeConfig                 `yaml:"runtime,omitempty"`
	EtcdAdvanced            *EtcdAdvancedConfig            `yaml:"etcdAdvanced,omitempty"`
	EmbeddedRegistry        *EmbeddedRegistryConfig        `yaml:"embeddedRegistry,omitempty"`
	ImageCredentialProvider *ImageCredentialProviderConfig `yaml:"imageCredentialProvider,omitempty"`
	LocalStorage            *LocalStorageConfig            `yaml:"localStorage,omitempty"`
	Agent                   *AgentConfig                   `yaml:"agent,omitempty"`
	VersionString           string                         `yaml:"-"`
}

func (k *K3s) SetDefaults() {
	if k.Config == nil {
		k.Config = dig.Mapping{}
	}
}

func (k *K3s) Validate() error {
	if k.Version != "" && !strings.HasPrefix(k.Version, "v") {
		return fmt.Errorf("k3s version must start with 'v': got %q", k.Version)
	}
	if es := k.EtcdSnapshots; es != nil {
		if es.S3 != nil && es.S3.Enabled {
			if es.S3.Bucket == "" {
				return fmt.Errorf("etcdSnapshots.s3.bucket is required when S3 is enabled")
			}
		}
	}
	if ag := k.Airgap; ag != nil && ag.Enabled {
		if ag.ImageBundle != "" {
			if !fileExistsLocal(ag.ImageBundle) {
				return fmt.Errorf("airgap.imageBundle: file not found: %s", ag.ImageBundle)
			}
		}
	}
	if se := k.SecretsEncrypt; se != nil {
		if (se.RotateKeys || se.Reencrypt) && !se.Enabled {
			return fmt.Errorf("secretsEncrypt.enabled must be true to use rotateKeys or reencrypt")
		}
	}
	if cr := k.CertRotate; cr != nil {
		if cr.RotateCA && !cr.Enabled {
			return fmt.Errorf("certRotate.enabled must be true to use rotateCA")
		}
	}
	if ep := k.EtcdSnapshots; ep != nil && ep.Prune != nil {
		if ep.Prune.Enabled && ep.Prune.Retention <= 0 {
			return fmt.Errorf("etcdSnapshots.prune.retention must be > 0 when pruning is enabled")
		}
	}
	if n := k.Network; n != nil {
		validFlannelBackends := map[string]bool{
			"none": true, "vxlan": true, "host-gw": true, "wireguard-native": true,
		}
		if n.FlannelBackend != "" && !validFlannelBackends[n.FlannelBackend] {
			return fmt.Errorf("network.flannelBackend: invalid value %q (valid: none, vxlan, host-gw, wireguard-native)", n.FlannelBackend)
		}
		validEgressModes := map[string]bool{
			"agent": true, "cluster": true, "pod": true, "disabled": true,
		}
		if n.EgressSelectorMode != "" && !validEgressModes[n.EgressSelectorMode] {
			return fmt.Errorf("network.egressSelectorMode: invalid value %q (valid: agent, cluster, pod, disabled)", n.EgressSelectorMode)
		}
	}
	if ds := k.Datastore; ds != nil {
		if ds.Endpoint == "" && (ds.CAFile != "" || ds.CertFile != "" || ds.KeyFile != "") {
			return fmt.Errorf("datastore.endpoint is required when TLS files are specified")
		}
	}
	if sec := k.Security; sec != nil {
		validProviders := map[string]bool{"aescbc": true, "secretbox": true}
		if sec.EncryptionProvider != "" && !validProviders[sec.EncryptionProvider] {
			return fmt.Errorf("security.encryptionProvider: invalid value %q (valid: aescbc, secretbox)", sec.EncryptionProvider)
		}
	}
	if n := k.Network; n != nil {
		if n.FlannelBackend == "none" && (n.FlannelIface != "" || n.FlannelConf != "" || n.FlannelCNIConf != "") {
			return fmt.Errorf("network: flannel interface/conf options are not valid when flannelBackend is \"none\"")
		}
	}
	if es := k.EtcdSnapshots; es != nil && es.S3 != nil {
		validLookupTypes := map[string]bool{"auto": true, "dns": true, "path": true}
		if es.S3.BucketLookupType != "" && !validLookupTypes[es.S3.BucketLookupType] {
			return fmt.Errorf("etcdSnapshots.s3.bucketLookupType: invalid value %q (valid: auto, dns, path)", es.S3.BucketLookupType)
		}
		if es.S3.SessionToken != "" && !es.S3.Enabled {
			return fmt.Errorf("etcdSnapshots.s3.sessionToken requires s3.enabled to be true")
		}
	}
	if icp := k.ImageCredentialProvider; icp != nil {
		if (icp.BinDir != "" && icp.Config == "") || (icp.BinDir == "" && icp.Config != "") {
			return fmt.Errorf("imageCredentialProvider: both binDir and config must be set together")
		}
	}
	if tok := k.Token; tok != nil {
		if tok.ServerToken != "" && tok.TokenFile != "" {
			return fmt.Errorf("token: serverToken and tokenFile are mutually exclusive")
		}
		if tok.AgentToken != "" && tok.AgentTokenFile != "" {
			return fmt.Errorf("token: agentToken and agentTokenFile are mutually exclusive")
		}
	}
	if ag := k.Agent; ag != nil {
		if ag.LBServerPort < 0 || ag.LBServerPort > 65535 {
			return fmt.Errorf("agent.lbServerPort: must be between 0 and 65535")
		}
	}
	return nil
}

func (k *K3s) GenerateServerConfig(host *Host, clusterToken string, isInit bool, joinAddr string) ([]byte, error) {
	config := make(dig.Mapping)

	for key, val := range k.Config {
		config[key] = val
	}

	if host.PrivateAddress != "" {
		config["node-ip"] = host.PrivateAddress
		config["advertise-address"] = host.PrivateAddress
	}

	if isInit {
		config["cluster-init"] = true
	} else if joinAddr != "" {
		config["server"] = fmt.Sprintf("https://%s:6443", joinAddr)
	}

	if clusterToken != "" {
		config["token"] = clusterToken
	}

	if host.NoFlannel {
		config["flannel-backend"] = "none"
	}

	applyNodeLabelsAndTaints(config, host)
	k.applyEtcdSnapshotConfig(config)
	k.applySecretsEncryptConfig(config)
	k.applyTokenConfig(config)
	k.applyNetworkConfig(config)
	k.applyDisableConfig(config)
	k.applyTLSConfig(config)
	k.applyKubeconfigConfig(config)
	k.applyDatastoreConfig(config)
	k.applyServerConfig(config)
	k.applySecurityConfig(config)
	k.applyComponentArgs(config, true)
	k.applyRuntimeConfig(config)
	k.applyEtcdAdvancedConfig(config)
	k.applyEmbeddedRegistryConfig(config)
	k.applyImageCredentialProviderConfig(config)
	k.applyLocalStorageConfig(config)
	applyNodeNetworking(config, host)

	for _, arg := range k.ExtraArgs {
		parts := strings.SplitN(arg, "=", 2)
		key := strings.TrimPrefix(parts[0], "--")
		if len(parts) == 2 {
			config[key] = parts[1]
		} else {
			config[key] = true
		}
	}

	return yaml.Marshal(config)
}

func (k *K3s) applyEtcdSnapshotConfig(config dig.Mapping) {
	es := k.EtcdSnapshots
	if es == nil {
		return
	}
	if es.Scheduled {
		if es.ScheduleCron != "" {
			config["etcd-snapshot-schedule-cron"] = es.ScheduleCron
		}
		if es.Retention > 0 {
			config["etcd-snapshot-retention"] = es.Retention
		}
		if es.Compress {
			config["etcd-snapshot-compress"] = true
		}
	}
	if es.Dir != "" {
		config["etcd-snapshot-dir"] = es.Dir
	}

	s3 := es.S3
	if s3 == nil || !s3.Enabled {
		return
	}
	config["etcd-s3"] = true
	if s3.Endpoint != "" {
		config["etcd-s3-endpoint"] = s3.Endpoint
	}
	if s3.EndpointCA != "" {
		config["etcd-s3-endpoint-ca"] = s3.EndpointCA
	}
	if s3.SkipSSLVerify {
		config["etcd-s3-skip-ssl-verify"] = true
	}
	if s3.AccessKey != "" {
		config["etcd-s3-access-key"] = s3.AccessKey
	}
	if s3.SecretKey != "" {
		config["etcd-s3-secret-key"] = s3.SecretKey
	}
	if s3.Bucket != "" {
		config["etcd-s3-bucket"] = s3.Bucket
	}
	if s3.Region != "" {
		config["etcd-s3-region"] = s3.Region
	}
	if s3.Folder != "" {
		config["etcd-s3-folder"] = s3.Folder
	}
	if s3.Insecure {
		config["etcd-s3-insecure"] = true
	}
	if s3.SessionToken != "" {
		config["etcd-s3-session-token"] = s3.SessionToken
	}
	if s3.BucketLookupType != "" {
		config["etcd-s3-bucket-lookup-type"] = s3.BucketLookupType
	}
	if s3.Proxy != "" {
		config["etcd-s3-proxy"] = s3.Proxy
	}
	if s3.ConfigSecret != "" {
		config["etcd-s3-config-secret"] = s3.ConfigSecret
	}
	if s3.Timeout != "" {
		config["etcd-s3-timeout"] = s3.Timeout
	}
	if s3.Retention > 0 {
		config["etcd-s3-retention"] = s3.Retention
	}
}

func (k *K3s) applySecretsEncryptConfig(config dig.Mapping) {
	if k.SecretsEncrypt == nil || !k.SecretsEncrypt.Enabled {
		return
	}
	config["secrets-encryption"] = true
}

func (k *K3s) applyTokenConfig(config dig.Mapping) {
	if k.Token == nil {
		return
	}
	if k.Token.AgentToken != "" {
		config["agent-token"] = k.Token.AgentToken
	}
	if k.Token.ServerToken != "" {
		config["token"] = k.Token.ServerToken
	}
	if k.Token.TokenFile != "" {
		config["token-file"] = k.Token.TokenFile
	}
	if k.Token.AgentTokenFile != "" {
		config["agent-token-file"] = k.Token.AgentTokenFile
	}
}

func (k *K3s) applyNetworkConfig(config dig.Mapping) {
	n := k.Network
	if n == nil {
		return
	}
	if len(n.ClusterCIDR) > 0 {
		config["cluster-cidr"] = strings.Join(n.ClusterCIDR, ",")
	}
	if len(n.ServiceCIDR) > 0 {
		config["service-cidr"] = strings.Join(n.ServiceCIDR, ",")
	}
	if len(n.ClusterDNS) > 0 {
		config["cluster-dns"] = strings.Join(n.ClusterDNS, ",")
	}
	if n.ClusterDomain != "" {
		config["cluster-domain"] = n.ClusterDomain
	}
	if n.NodePortRange != "" {
		config["service-node-port-range"] = n.NodePortRange
	}
	if n.FlannelBackend != "" {
		config["flannel-backend"] = n.FlannelBackend
	}
	if n.FlannelIPv6Masq {
		config["flannel-ipv6-masq"] = true
	}
	if n.FlannelExternalIP {
		config["flannel-external-ip"] = true
	}
	if n.FlannelIface != "" {
		config["flannel-iface"] = n.FlannelIface
	}
	if n.FlannelConf != "" {
		config["flannel-conf"] = n.FlannelConf
	}
	if n.FlannelCNIConf != "" {
		config["flannel-cni-conf"] = n.FlannelCNIConf
	}
	if n.EgressSelectorMode != "" {
		config["egress-selector-mode"] = n.EgressSelectorMode
	}
	if n.ServiceLBNamespace != "" {
		config["servicelb-namespace"] = n.ServiceLBNamespace
	}
}

func (k *K3s) applyDisableConfig(config dig.Mapping) {
	d := k.Disable
	if d == nil {
		return
	}
	if len(d.Components) > 0 {
		config["disable"] = d.Components
	}
	if d.Scheduler {
		config["disable-scheduler"] = true
	}
	if d.CloudController {
		config["disable-cloud-controller"] = true
	}
	if d.KubeProxy {
		config["disable-kube-proxy"] = true
	}
	if d.NetworkPolicy {
		config["disable-network-policy"] = true
	}
	if d.HelmController {
		config["disable-helm-controller"] = true
	}
}

func (k *K3s) applyTLSConfig(config dig.Mapping) {
	t := k.TLS
	if t == nil {
		return
	}
	if len(t.SANs) > 0 {
		config["tls-san"] = t.SANs
	}
	if t.SANSecurity {
		config["tls-san-security"] = true
	}
}

func (k *K3s) applyKubeconfigConfig(config dig.Mapping) {
	kc := k.Kubeconfig
	if kc == nil {
		return
	}
	if kc.Mode != "" {
		config["write-kubeconfig-mode"] = kc.Mode
	}
	if kc.Group != "" {
		config["write-kubeconfig-group"] = kc.Group
	}
	if kc.Name != "" {
		config["write-kubeconfig-name"] = kc.Name
	}
	if kc.Path != "" {
		config["write-kubeconfig"] = kc.Path
	}
}

func (k *K3s) applyDatastoreConfig(config dig.Mapping) {
	ds := k.Datastore
	if ds == nil {
		return
	}
	if ds.Endpoint != "" {
		config["datastore-endpoint"] = ds.Endpoint
	}
	if ds.CAFile != "" {
		config["datastore-cafile"] = ds.CAFile
	}
	if ds.CertFile != "" {
		config["datastore-certfile"] = ds.CertFile
	}
	if ds.KeyFile != "" {
		config["datastore-keyfile"] = ds.KeyFile
	}
}

func (k *K3s) applyServerConfig(config dig.Mapping) {
	s := k.Server
	if s == nil {
		return
	}
	if s.AdvertiseAddress != "" {
		config["advertise-address"] = s.AdvertiseAddress
	}
	if s.AdvertisePort > 0 {
		config["advertise-port"] = s.AdvertisePort
	}
	if s.BindAddress != "" {
		config["bind-address"] = s.BindAddress
	}
	if s.HTTPSPort > 0 {
		config["https-listen-port"] = s.HTTPSPort
	}
	if s.SupervisorMetrics {
		config["supervisor-metrics"] = true
	}
	if s.EnablePprof {
		config["enable-pprof"] = true
	}
}

func (k *K3s) applySecurityConfig(config dig.Mapping) {
	sec := k.Security
	if sec == nil {
		return
	}
	if sec.SELinux {
		config["selinux"] = true
	}
	if sec.ProtectKernelDefaults {
		config["protect-kernel-defaults"] = true
	}
	if sec.EncryptionProvider != "" {
		config["secrets-encryption-provider"] = sec.EncryptionProvider
	}
	if sec.Rootless {
		config["rootless"] = true
	}
	if sec.PreferBundledBin {
		config["prefer-bundled-bin"] = true
	}
}

func (k *K3s) applyComponentArgs(config dig.Mapping, isServer bool) {
	ca := k.ComponentArgs
	if ca == nil {
		return
	}
	if isServer {
		if len(ca.KubeAPIServer) > 0 {
			config["kube-apiserver-arg"] = ca.KubeAPIServer
		}
		if len(ca.KubeControllerManager) > 0 {
			config["kube-controller-manager-arg"] = ca.KubeControllerManager
		}
		if len(ca.KubeScheduler) > 0 {
			config["kube-scheduler-arg"] = ca.KubeScheduler
		}
		if len(ca.KubeCloudControllerManager) > 0 {
			config["kube-cloud-controller-manager-arg"] = ca.KubeCloudControllerManager
		}
		if len(ca.Etcd) > 0 {
			config["etcd-arg"] = ca.Etcd
		}
		if len(ca.HelmController) > 0 {
			config["helm-controller-arg"] = ca.HelmController
		}
	}
	if len(ca.Kubelet) > 0 {
		config["kubelet-arg"] = ca.Kubelet
	}
	if len(ca.KubeProxy) > 0 {
		config["kube-proxy-arg"] = ca.KubeProxy
	}
}

func (k *K3s) applyRuntimeConfig(config dig.Mapping) {
	rt := k.Runtime
	if rt == nil {
		return
	}
	if rt.SystemDefaultRegistry != "" {
		config["system-default-registry"] = rt.SystemDefaultRegistry
	}
	if rt.PauseImage != "" {
		config["pause-image"] = rt.PauseImage
	}
	if rt.Snapshotter != "" {
		config["snapshotter"] = rt.Snapshotter
	}
	if rt.DefaultRuntime != "" {
		config["default-runtime"] = rt.DefaultRuntime
	}
	if rt.ResolvConf != "" {
		config["resolv-conf"] = rt.ResolvConf
	}
	if rt.ContainerRuntimeEndpoint != "" {
		config["container-runtime-endpoint"] = rt.ContainerRuntimeEndpoint
	}
	if rt.ImageServiceEndpoint != "" {
		config["image-service-endpoint"] = rt.ImageServiceEndpoint
	}
}

func (k *K3s) applyEtcdAdvancedConfig(config dig.Mapping) {
	ea := k.EtcdAdvanced
	if ea == nil {
		return
	}
	if ea.ExposeMetrics {
		config["etcd-expose-metrics"] = true
	}
	if ea.DisableSnapshots {
		config["etcd-disable-snapshots"] = true
	}
	if ea.SnapshotName != "" {
		config["etcd-snapshot-name"] = ea.SnapshotName
	}
	if ea.SnapshotReconcileInterval != "" {
		config["etcd-snapshot-reconcile-interval"] = ea.SnapshotReconcileInterval
	}
}

func (k *K3s) applyEmbeddedRegistryConfig(config dig.Mapping) {
	er := k.EmbeddedRegistry
	if er == nil {
		return
	}
	if er.Enabled {
		config["embedded-registry"] = true
	}
	if er.DisableDefaultRegistryEndpoint {
		config["disable-default-registry-endpoint"] = true
	}
}

func (k *K3s) applyImageCredentialProviderConfig(config dig.Mapping) {
	icp := k.ImageCredentialProvider
	if icp == nil {
		return
	}
	if icp.BinDir != "" {
		config["image-credential-provider-bin-dir"] = icp.BinDir
	}
	if icp.Config != "" {
		config["image-credential-provider-config"] = icp.Config
	}
}

func (k *K3s) applyLocalStorageConfig(config dig.Mapping) {
	ls := k.LocalStorage
	if ls == nil {
		return
	}
	if ls.Path != "" {
		config["default-local-storage-path"] = ls.Path
	}
}

func (k *K3s) applyDisableDefaultRegistryEndpoint(config dig.Mapping) {
	er := k.EmbeddedRegistry
	if er == nil {
		return
	}
	if er.DisableDefaultRegistryEndpoint {
		config["disable-default-registry-endpoint"] = true
	}
}

func (k *K3s) applyAgentConfig(config dig.Mapping) {
	ag := k.Agent
	if ag == nil {
		return
	}
	if ag.LBServerPort > 0 {
		config["lb-server-port"] = ag.LBServerPort
	}
}

func applyNodeNetworking(config dig.Mapping, host *Host) {
	if len(host.NodeExternalIP) > 0 {
		config["node-external-ip"] = strings.Join(host.NodeExternalIP, ",")
	}
	if len(host.NodeInternalDNS) > 0 {
		config["node-internal-dns"] = strings.Join(host.NodeInternalDNS, ",")
	}
	if len(host.NodeExternalDNS) > 0 {
		config["node-external-dns"] = strings.Join(host.NodeExternalDNS, ",")
	}
	if host.WithNodeID {
		config["with-node-id"] = true
	}
}

func (k *K3s) GenerateAgentConfig(host *Host, serverAddr, clusterToken string) ([]byte, error) {
	config := make(dig.Mapping)

	for key, val := range k.Config {
		switch key {
		case "cluster-init", "cluster-cidr", "service-cidr", "cluster-dns",
			"disable", "tls-san", "etcd-expose-metrics",
			"write-kubeconfig-mode", "kube-apiserver-arg",
			"kube-controller-manager-arg", "kube-scheduler-arg":
			continue
		default:
			config[key] = val
		}
	}

	if serverAddr != "" {
		config["server"] = fmt.Sprintf("https://%s:6443", serverAddr)
	}
	if clusterToken != "" {
		config["token"] = clusterToken
	}

	if host.PrivateAddress != "" {
		config["node-ip"] = host.PrivateAddress
	}

	applyNodeLabelsAndTaints(config, host)
	k.applySecurityConfig(config)
	k.applyComponentArgs(config, false)
	k.applyRuntimeConfig(config)
	k.applyImageCredentialProviderConfig(config)
	k.applyAgentConfig(config)
	k.applyDisableDefaultRegistryEndpoint(config)
	applyNodeNetworking(config, host)

	return yaml.Marshal(config)
}

func applyNodeLabelsAndTaints(config dig.Mapping, host *Host) {
	if len(host.Labels) > 0 {
		labels := make([]string, 0, len(host.Labels))
		for k, v := range host.Labels {
			labels = append(labels, fmt.Sprintf("%s=%s", k, v))
		}
		config["node-label"] = labels
	}
	if len(host.Taints) > 0 {
		config["node-taint"] = host.Taints
	}
}
