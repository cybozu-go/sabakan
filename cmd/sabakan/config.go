package main

const (
	defaultListenHTTP  = "0.0.0.0:10080"
	defaultEtcdPrefix  = "/sabakan/"
	defaultEtcdTimeout = "2s"
	defaultDHCPBind    = "0.0.0.0:10067"
	defaultIPXEPath    = "/usr/lib/ipxe/ipxe.efi"
	defaultDataDir     = "/var/lib/sabakan"
)

var (
	defaultEtcdServers = []string{"http://localhost:2379"}
	defaultAllowIPs    = []string{"127.0.0.1", "::1"}
)

func newConfig() *config {
	return &config{
		ListenHTTP:  defaultListenHTTP,
		EtcdServers: defaultEtcdServers,
		EtcdPrefix:  defaultEtcdPrefix,
		EtcdTimeout: defaultEtcdTimeout,
		DHCPBind:    defaultDHCPBind,
		IPXEPath:    defaultIPXEPath,
		DataDir:     defaultDataDir,
		AllowIPs:    defaultAllowIPs,
	}
}

type config struct {
	ListenHTTP   string   `yaml:"http"`
	EtcdServers  []string `yaml:"etcd-servers"`
	EtcdPrefix   string   `yaml:"etcd-prefix"`
	EtcdTimeout  string   `yaml:"etcd-timeout"`
	EtcdUsername string   `yaml:"etcd-username"`
	EtcdPassword string   `yaml:"etcd-password"`
	EtcdTLS      bool     `yaml:"etcd-tls"`
	EtcdTLSCA    string   `yaml:"etcd-tls-ca"`
	EtcdTLSCert  string   `yaml:"etcd-tls-cert"`
	EtcdTLSKey   string   `yaml:"etcd-tls-key"`
	DHCPBind     string   `yaml:"dhcp-bind"`
	IPXEPath     string   `yaml:"ipxe-efi-path"`
	DataDir      string   `yaml:"data-dir"`
	AdvertiseURL string   `yaml:"advertise-url"`
	AllowIPs     []string `yaml:"allow-ips"`
}
