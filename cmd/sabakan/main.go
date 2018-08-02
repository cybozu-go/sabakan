package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/namespace"
	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan/dhcpd"
	"github.com/cybozu-go/sabakan/models/etcd"
	"github.com/cybozu-go/sabakan/web"
	"go.universe.tf/netboot/dhcp4"
	yaml "gopkg.in/yaml.v2"
)

var (
	flagHTTP         = flag.String("http", defaultListenHTTP, "<Listen IP>:<Port number>")
	flagEtcdServers  = flag.String("etcd-servers", strings.Join(defaultEtcdServers, ","), "comma-separated URLs of the backend etcd")
	flagEtcdPrefix   = flag.String("etcd-prefix", defaultEtcdPrefix, "etcd prefix")
	flagEtcdTimeout  = flag.String("etcd-timeout", "2s", "dial timeout to etcd")
	flagEtcdUsername = flag.String("etcd-username", "", "username for etcd authentication")
	flagEtcdPassword = flag.String("etcd-password", "", "password for etcd authentication")
	flagEtcdTLSCA    = flag.String("etcd-tls-ca", "", "path to CA bundle used to verify certificates of etcd servers")
	flagEtcdTLSCert  = flag.String("etcd-tls-cert", "", "path to my certificate used to identify myself to etcd servers")
	flagEtcdTLSKey   = flag.String("etcd-tls-key", "", "path to my key used to identify myself to etcd servers")

	flagDHCPBind     = flag.String("dhcp-bind", defaultDHCPBind, "bound ip addresses and port for dhcp server")
	flagIPXEPath     = flag.String("ipxe-efi-path", defaultIPXEPath, "path to ipxe.efi")
	flagDataDir      = flag.String("data-dir", defaultDataDir, "directory to store files")
	flagAdvertiseURL = flag.String("advertise-url", "", "public URL of this server")
	flagAllowIPs     = flag.String("allow-ips", strings.Join(defaultAllowIPs, ","), "comma-separated IPs allowed to change resources")

	flagConfigFile = flag.String("config-file", "", "path to configuration file")
)

func main() {
	flag.Parse()
	cmd.LogConfig{}.Apply()

	// seed math/random
	rand.Seed(time.Now().UnixNano())

	cfg := newConfig()
	if *flagConfigFile == "" {
		cfg.ListenHTTP = *flagHTTP
		cfg.EtcdServers = strings.Split(*flagEtcdServers, ",")
		cfg.EtcdPrefix = *flagEtcdPrefix
		cfg.EtcdTimeout = *flagEtcdTimeout
		cfg.EtcdUsername = *flagEtcdUsername
		cfg.EtcdPassword = *flagEtcdPassword
		cfg.EtcdTLSCA = *flagEtcdTLSCA
		cfg.EtcdTLSCert = *flagEtcdTLSCert
		cfg.EtcdTLSKey = *flagEtcdTLSKey
		cfg.DHCPBind = *flagDHCPBind
		cfg.IPXEPath = *flagIPXEPath
		cfg.DataDir = *flagDataDir
		cfg.AdvertiseURL = *flagAdvertiseURL
		cfg.AllowIPs = strings.Split(*flagAllowIPs, ",")
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

	if !filepath.IsAbs(cfg.DataDir) {
		fmt.Fprintln(os.Stderr, "data-dir must be an absolute path")
		os.Exit(1)
	}
	if cfg.AdvertiseURL == "" {
		fmt.Fprintln(os.Stderr, "advertise-url must be specified")
		os.Exit(1)
	}
	advertiseURL, err := url.Parse(cfg.AdvertiseURL)
	if err != nil {
		log.ErrorExit(err)
	}

	timeout, err := time.ParseDuration(cfg.EtcdTimeout)
	if err != nil {
		log.ErrorExit(err)
	}

	etcdCfg := clientv3.Config{
		Endpoints:   cfg.EtcdServers,
		DialTimeout: timeout,
		Username:    cfg.EtcdUsername,
		Password:    cfg.EtcdPassword,
	}

	tlsCfg := &tls.Config{}
	if len(cfg.EtcdTLSCA) != 0 {
		rootCACert, err := ioutil.ReadFile(cfg.EtcdTLSCA)
		if err != nil {
			log.ErrorExit(err)
		}
		rootCAs := x509.NewCertPool()
		ok := rootCAs.AppendCertsFromPEM(rootCACert)
		if !ok {
			fmt.Fprintln(os.Stderr, "Failed to parse PEM file")
			os.Exit(1)
		}
		tlsCfg.RootCAs = rootCAs
	}
	if len(cfg.EtcdTLSCert) != 0 && len(cfg.EtcdTLSKey) != 0 {
		cert, err := tls.LoadX509KeyPair(cfg.EtcdTLSCert, cfg.EtcdTLSKey)
		if err != nil {
			log.ErrorExit(err)
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}
	etcdCfg.TLS = tlsCfg

	c, err := clientv3.New(etcdCfg)
	if err != nil {
		log.ErrorExit(err)
	}
	c.KV = namespace.NewKV(c.KV, cfg.EtcdPrefix)
	c.Watcher = namespace.NewWatcher(c.Watcher, cfg.EtcdPrefix)
	c.Lease = namespace.NewLease(c.Lease, cfg.EtcdPrefix)
	defer c.Close()

	model := etcd.NewModel(c, cfg.DataDir, advertiseURL)
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
		Handler: dhcpd.DHCPHandler{Model: model, MyURL: advertiseURL},
		Conn:    conn,
	}
	cmd.Go(dhcpServer.Serve)

	allowedIPs, err := parseAllowIPs(cfg.AllowIPs)
	if err != nil {
		log.ErrorExit(err)
	}
	webServer := web.Server{
		Model:          model,
		IPXEFirmware:   cfg.IPXEPath,
		MyURL:          advertiseURL,
		AllowedRemotes: allowedIPs,
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

func parseAllowIPs(ips []string) ([]*net.IPNet, error) {
	nets := make([]*net.IPNet, len(ips))
	for i, cidr := range ips {
		if !strings.Contains(cidr, "/") {
			cidr += "/32"
		}
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}
		nets[i] = ipnet
	}
	return nets, nil
}
