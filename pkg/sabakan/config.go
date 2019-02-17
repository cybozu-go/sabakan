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
		Etcd:       etcdutil.NewConfig(defaultEtcdPrefix),
	}
}

type config struct {
	ListenHTTP   string           `json:"http"`
	DHCPBind     string           `json:"dhcp-bind"`
	IPXEPath     string           `json:"ipxe-efi-path"`
	DataDir      string           `json:"data-dir"`
	AdvertiseURL string           `json:"advertise-url"`
	AllowIPs     []string         `json:"allow-ips"`
	Playground   bool             `json:"enable-playground"`
	Etcd         *etcdutil.Config `json:"etcd"`
}
