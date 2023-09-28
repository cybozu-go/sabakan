package main

import "github.com/cybozu-go/etcdutil"

const (
	defaultListenHTTP         = "0.0.0.0:10080"
	defaultListenHTTPS        = "0.0.0.0:10443"
	defaultListenMetrics      = "0.0.0.0:10081"
	defaultEtcdPrefix         = "/sabakan/"
	defaultDHCPBind           = "0.0.0.0:10067"
	defaultIPXEPath           = "/usr/lib/ipxe/ipxe.efi"
	defaultDataDir            = "/var/lib/sabakan"
	defaultSabakanTLSCertFile = "/etc/sabakan/sabakan-tls.crt"
	defaultSabakanTLSKeyFile  = "/etc/sabakan/sabakan-tls.key"
)

var (
	defaultAllowIPs = []string{"127.0.0.1", "::1"}
)

func newConfig() *config {
	return &config{
		ListenHTTP:         defaultListenHTTP,
		ListenHTTPS:        defaultListenHTTPS,
		ListenMetrics:      defaultListenMetrics,
		DHCPBind:           defaultDHCPBind,
		IPXEPath:           defaultIPXEPath,
		DataDir:            defaultDataDir,
		AllowIPs:           defaultAllowIPs,
		Etcd:               etcdutil.NewConfig(defaultEtcdPrefix),
		SabakanTLSCertFile: defaultSabakanTLSCertFile,
		SabakanTLSKeyFile:  defaultSabakanTLSKeyFile,
	}
}

type config struct {
	ListenHTTP        string `json:"http"`
	ListenHTTPS       string `json:"https"`
	ListenMetrics     string `json:"metrics"`
	DHCPBind          string `json:"dhcp-bind"`
	IPXEPath          string `json:"ipxe-efi-path"`
	DataDir           string `json:"data-dir"`
	AdvertiseURL      string `json:"advertise-url"`
	AdvertiseURLHTTPS string `json:"advertise-url-https"`

	AllowIPs           []string         `json:"allow-ips"`
	Playground         bool             `json:"enable-playground"`
	Etcd               *etcdutil.Config `json:"etcd"`
	SabakanTLSCertFile string           `json:"sabakan-tls-cert"`
	SabakanTLSKeyFile  string           `json:"sabakan-tls-key"`
}
