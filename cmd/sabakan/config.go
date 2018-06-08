package main

const (
	defaultListenHTTP  = "0.0.0.0:10080"
	defaultEtcdPrefix  = "/sabakan"
	defaultEtcdTimeout = "2s"
	defaultDHCPBind    = "0.0.0.0:10067"
	defaultIPXEPath    = "/usr/lib/ipxe/ipxe.efi"
	defaultImageDir    = "/var/lib/sabakan"
)

var (
	defaultEtcdServers = []string{"http://localhost:2379"}
	defaultAllowIPs    = []string{"127.0.0.1"}
)

func newConfig() *config {
	return &config{
		ListenHTTP:  defaultListenHTTP,
		EtcdServers: defaultEtcdServers,
		EtcdPrefix:  defaultEtcdPrefix,
		EtcdTimeout: defaultEtcdTimeout,
		DHCPBind:    defaultDHCPBind,
		IPXEPath:    defaultIPXEPath,
		ImageDir:    defaultImageDir,
		AllowIPs:    defaultAllowIPs,
	}
}

type config struct {
	ListenHTTP   string   `yaml:"http"`
	EtcdServers  []string `yaml:"etcd-servers"`
	EtcdPrefix   string   `yaml:"etcd-prefix"`
	EtcdTimeout  string   `yaml:"etcd-timeout"`
	DHCPBind     string   `yaml:"dhcp-bind"`
	IPXEPath     string   `yaml:"ipxe-efi-path"`
	ImageDir     string   `yaml:"image-dir"`
	AdvertiseURL string   `yaml:"advertise-url"`
	AllowIPs     []string `yaml:"allow-ips"`
}
