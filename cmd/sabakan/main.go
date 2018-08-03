package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/etcdutil"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan/dhcpd"
	"github.com/cybozu-go/sabakan/models/etcd"
	"github.com/cybozu-go/sabakan/web"
	"go.universe.tf/netboot/dhcp4"
	yaml "gopkg.in/yaml.v2"
)

var (
	flagHTTP         = flag.String("http", defaultListenHTTP, "<Listen IP>:<Port number>")
	flagDHCPBind     = flag.String("dhcp-bind", defaultDHCPBind, "bound ip addresses and port for dhcp server")
	flagIPXEPath     = flag.String("ipxe-efi-path", defaultIPXEPath, "path to ipxe.efi")
	flagDataDir      = flag.String("data-dir", defaultDataDir, "directory to store files")
	flagAdvertiseURL = flag.String("advertise-url", "", "public URL of this server")
	flagAllowIPs     = flag.String("allow-ips", strings.Join(defaultAllowIPs, ","), "comma-separated IPs allowed to change resources")

	flagEtcdEndpoints = flag.String("etcd-endpoints", strings.Join(etcdutil.DefaultEndpoints, ","), "comma-separated URLs of the backend etcd endpoints")
	flagEtcdPrefix    = flag.String("etcd-prefix", defaultEtcdPrefix, "etcd prefix")
	flagEtcdTimeout   = flag.String("etcd-timeout", etcdutil.DefaultTimeout, "dial timeout to etcd")
	flagEtcdUsername  = flag.String("etcd-username", "", "username for etcd authentication")
	flagEtcdPassword  = flag.String("etcd-password", "", "password for etcd authentication")
	flagEtcdTLSCA     = flag.String("etcd-tls-ca", "", "path to CA bundle used to verify certificates of etcd servers")
	flagEtcdTLSCert   = flag.String("etcd-tls-cert", "", "path to my certificate used to identify myself to etcd servers")
	flagEtcdTLSKey    = flag.String("etcd-tls-key", "", "path to my key used to identify myself to etcd servers")

	flagConfigFile = flag.String("config-file", "", "path to configuration file")
)

func main() {
	flag.Parse()
	cmd.LogConfig{}.Apply()

	// seed math/random
	rand.Seed(time.Now().UnixNano())

	cfg := newConfig()
	if *flagConfigFile == "" {
		cfg.AdvertiseURL = *flagAdvertiseURL
		cfg.AllowIPs = strings.Split(*flagAllowIPs, ",")
		cfg.DHCPBind = *flagDHCPBind
		cfg.DataDir = *flagDataDir
		cfg.IPXEPath = *flagIPXEPath
		cfg.ListenHTTP = *flagHTTP

		cfg.Etcd.Endpoints = strings.Split(*flagEtcdEndpoints, ",")
		cfg.Etcd.Prefix = *flagEtcdPrefix
		cfg.Etcd.Timeout = *flagEtcdTimeout
		cfg.Etcd.Username = *flagEtcdUsername
		cfg.Etcd.Password = *flagEtcdPassword
		cfg.Etcd.TLSCA = *flagEtcdTLSCA
		cfg.Etcd.TLSCert = *flagEtcdTLSCert
		cfg.Etcd.TLSKey = *flagEtcdTLSKey
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

	c, err := cfg.Etcd.Client()
	if err != nil {
		log.ErrorExit(err)
	}
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
