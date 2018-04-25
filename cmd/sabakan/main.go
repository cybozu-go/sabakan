package main

import (
	"context"
	"flag"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan"
	dhcp "github.com/cybozu-go/sabakan/dhcp"
	"github.com/gorilla/mux"
)

var (
	flagHTTP        = flag.String("http", "0.0.0.0:8888", "<Listen IP>:<Port number>")
	flagEtcdServers = flag.String("etcd-servers", "http://localhost:2379", "URLs of the backend etcd")
	flagEtcdPrefix  = flag.String("etcd-prefix", "", "etcd prefix")
	flagEtcdTimeout = flag.String("etcd-timeout", "2s", "dial timeout to etcd")

	flagDHCPBind         = flag.String("dhcp-bind", "0.0.0.0:67", "bound ip addresses and port for dhcp server")
	flagDHCPInterface    = flag.String("dhcp-interface", "", "interface which receive a packet on")
	flagDHCPIPXEFirmware = flag.String("dhcp-ipxe-firmware-url", "", "URL to iPXE firmware")
)

var dhcpBegin = net.IPv4(10, 69, 0, 33)
var dhcpEnd = net.IPv4(10, 69, 0, 63)

func main() {
	flag.Parse()

	var e sabakan.EtcdConfig
	e.Servers = strings.Split(*flagEtcdServers, ",")
	e.Prefix = "/" + *flagEtcdPrefix

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

	ctx := context.Background()
	err = sabakan.Indexing(ctx, c, e.Prefix)
	if err != nil {
		log.ErrorExit(err)
	}

	r := mux.NewRouter()
	etcdClient := &sabakan.EtcdClient{Client: c, Prefix: e.Prefix}
	sabakan.InitConfig(r.PathPrefix("/api/v1/").Subrouter(), etcdClient)
	sabakan.InitCrypts(r.PathPrefix("/api/v1/").Subrouter(), etcdClient)
	sabakan.InitMachines(r.PathPrefix("/api/v1/").Subrouter(), etcdClient)

	cmd.Go(func(ctx context.Context) error {
		return sabakan.EtcdWatcher(ctx, e)
	})

	dhcps := dhcp.New(*flagDHCPBind, *flagDHCPInterface, *flagDHCPIPXEFirmware, dhcpBegin, dhcpEnd)
	cmd.Go(func(ctx context.Context) error {
		return dhcps.Serve(ctx)
	})

	s := &cmd.HTTPServer{
		Server: &http.Server{
			Addr:    *flagHTTP,
			Handler: r,
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
