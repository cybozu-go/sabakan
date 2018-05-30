package main

const (
	defaultListenHTTP  = "0.0.0.0:10080"
	defaultURLPort     = "10080"
	defaultEtcdPrefix  = "/sabakan"
	defaultEtcdTimeout = "2s"
	defaultDHCPBind    = "0.0.0.0:10067"
	defaultIPXEPath    = "/usr/lib/ipxe/ipxe.efi"
	defaultImageDir    = "/var/lib/sabakan"
)

var (
	defaultEtcdServers = []string{"http://localhost:2379"}
)

func newConfig() *config {
	return &config{
		ListenHTTP:  defaultListenHTTP,
		URLPort:     defaultURLPort,
		EtcdServers: defaultEtcdServers,
		EtcdPrefix:  defaultEtcdPrefix,
		EtcdTimeout: defaultEtcdTimeout,
		DHCPBind:    defaultDHCPBind,
		IPXEPath:    defaultIPXEPath,
		ImageDir:    defaultImageDir,
	}
}

type config struct {
	ListenHTTP  string   `yaml:"http"`
	URLPort     string   `yaml:"url-port"`
	EtcdServers []string `yaml:"etcd-servers"`
	EtcdPrefix  string   `yaml:"etcd-prefix"`
	EtcdTimeout string   `yaml:"etcd-timeout"`
	DHCPBind    string   `yaml:"dhcp-bind"`
	IPXEPath    string   `yaml:"ipxe-efi-path"`
	ImageDir    string   `yaml:"image-dir"`
}
