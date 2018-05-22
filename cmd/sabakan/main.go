package main

import (
	"context"
	"flag"
	"net/http"
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
)

type etcdConfig struct {
	Servers []string
	Prefix  string
}

var (
	flagHTTP        = flag.String("http", "0.0.0.0:10080", "<Listen IP>:<Port number>")
	flagEtcdServers = flag.String("etcd-servers", "http://localhost:2379", "comma-separated URLs of the backend etcd")
	flagEtcdPrefix  = flag.String("etcd-prefix", "/sabakan", "etcd prefix")
	flagEtcdTimeout = flag.String("etcd-timeout", "2s", "dial timeout to etcd")

	flagDHCPBind         = flag.String("dhcp-bind", "0.0.0.0:10067", "bound ip addresses and port for dhcp server")
	flagDHCPIPXEFirmware = flag.String("dhcp-ipxe-firmware-url", "", "URL to iPXE firmware")
)

func main() {
	flag.Parse()

	var e etcdConfig
	e.Servers = strings.Split(*flagEtcdServers, ",")
	e.Prefix = path.Clean("/" + *flagEtcdPrefix)

	timeout, err := time.ParseDuration(*flagEtcdTimeout)
	if err != nil {
		log.ErrorExit(err)
	}

	cfg := clientv3.Config{
		Endpoints:   e.Servers,
		DialTimeout: timeout,
	}
	c, err := clientv3.New(cfg)
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

	conn, err := dhcp4.NewConn(*flagDHCPBind)
	if err != nil {
		log.ErrorExit(err)
	}
	defer conn.Close()
	dhcpServer := dhcpd.Server{
		Handler: dhcpd.DHCPHandler{Model: model},
		Conn:    conn,
	}
	cmd.Go(dhcpServer.Serve)

	s := &cmd.HTTPServer{
		Server: &http.Server{
			Addr:    *flagHTTP,
			Handler: web.Server{model},
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
