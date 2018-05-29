package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan/dhcpd"
	"github.com/cybozu-go/sabakan/models/etcd"
	"github.com/cybozu-go/sabakan/web"
	"go.universe.tf/netboot/dhcp4"
	yaml "gopkg.in/yaml.v2"
)

type etcdConfig struct {
	Servers []string
	Prefix  string
}

var (
	flagHTTP        = flag.String("http", "0.0.0.0:10080", "<Listen IP>:<Port number>")
	flagURLPort     = flag.String("url-port", "10080", "port number used to construct boot API URL")
	flagEtcdServers = flag.String("etcd-servers", "http://localhost:2379", "comma-separated URLs of the backend etcd")
	flagEtcdPrefix  = flag.String("etcd-prefix", "/sabakan", "etcd prefix")
	flagEtcdTimeout = flag.String("etcd-timeout", "2s", "dial timeout to etcd")

	flagDHCPBind = flag.String("dhcp-bind", "0.0.0.0:10067", "bound ip addresses and port for dhcp server")
	flagIPXEPath = flag.String("ipxe-efi-path", "/usr/lib/ipxe/ipxe.efi", "path to ipxe.efi")

	flagConfigFile = flag.String("config-file", "", "path to configuration file")
)

func main() {
	flag.Parse()
	cmd.LogConfig{}.Apply()

	cfg := newConfig()
	if *flagConfigFile == "" {
		cfg.ListenHTTP = *flagHTTP
		cfg.URLPort = *flagURLPort
		cfg.EtcdServers = strings.Split(*flagEtcdServers, ",")
		cfg.EtcdPrefix = *flagEtcdPrefix
		cfg.EtcdTimeout = *flagEtcdTimeout
		cfg.DHCPBind = *flagDHCPBind
		cfg.IPXEPath = *flagIPXEPath
	} else {
		f, err := os.Open(*flagConfigFile)
		if err != nil {
			log.ErrorExit(err)
		}
		err = yaml.NewDecoder(f).Decode(cfg)
		if err != nil {
			log.ErrorExit(err)
		}
		f.Close()
	}

	var e etcdConfig
	e.Servers = cfg.EtcdServers
	e.Prefix = path.Clean("/" + cfg.EtcdPrefix)

	timeout, err := time.ParseDuration(cfg.EtcdTimeout)
	if err != nil {
		log.ErrorExit(err)
	}

	etcdCfg := clientv3.Config{
		Endpoints:   e.Servers,
		DialTimeout: timeout,
	}
	c, err := clientv3.New(etcdCfg)
	if err != nil {
		log.ErrorExit(err)
	}
	defer c.Close()

	model := etcd.NewModel(c, e.Prefix)
	ch := make(chan struct{})
	cmd.Go(func(ctx context.Context) error {
		return model.Run(ctx, ch)
	})
	// waiting the driver gets ready
	<-ch

	conn, err := dhcp4.NewConn(cfg.DHCPBind)
	if err != nil {
		log.ErrorExit(err)
	}
	dhcpServer := dhcpd.Server{
		Handler: dhcpd.DHCPHandler{Model: model, URLPort: cfg.URLPort},
		Conn:    conn,
	}
	cmd.Go(dhcpServer.Serve)

	webServer := web.Server{
		Model:        model,
		IPXEFirmware: cfg.IPXEPath,
	}
	s := &cmd.HTTPServer{
		Server: &http.Server{
			Addr:    cfg.ListenHTTP,
			Handler: webServer,
		},
		ShutdownTimeout: 3 * time.Minute,
	}
	s.ListenAndServe()

	cmd.Stop()
	err = cmd.Wait()
	if !cmd.IsSignaled(err) && err != nil {
		log.ErrorExit(err)
	}
}
