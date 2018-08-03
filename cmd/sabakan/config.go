package main

import "github.com/cybozu-go/etcdutil"

const (
	defaultListenHTTP = "0.0.0.0:10080"
	defaultEtcdPrefix = "/sabakan/"
	defaultDHCPBind   = "0.0.0.0:10067"
	defaultIPXEPath   = "/usr/lib/ipxe/ipxe.efi"
	defaultDataDir    = "/var/lib/sabakan"
)

var (
	defaultAllowIPs = []string{"127.0.0.1", "::1"}
)

func newConfig() *config {
	return &config{
		ListenHTTP: defaultListenHTTP,
		DHCPBind:   defaultDHCPBind,
		IPXEPath:   defaultIPXEPath,
		DataDir:    defaultDataDir,
		AllowIPs:   defaultAllowIPs,
		Etcd:       etcdutil.NewConfig(),
	}
}

type config struct {
	ListenHTTP   string           `yaml:"http"`
	DHCPBind     string           `yaml:"dhcp-bind"`
	IPXEPath     string           `yaml:"ipxe-efi-path"`
	DataDir      string           `yaml:"data-dir"`
	AdvertiseURL string           `yaml:"advertise-url"`
	AllowIPs     []string         `yaml:"allow-ips"`
	Etcd         *etcdutil.Config `yaml:"etcd"`
}
