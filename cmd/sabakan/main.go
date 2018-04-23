package main

import (
	"context"
	"flag"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan"
	"github.com/gorilla/mux"
)

var (
	flagHTTP        = flag.String("http", "0.0.0.0:8888", "<Listen IP>:<Port number>")
	flagEtcdServers = flag.String("etcd-servers", "http://localhost:2379", "URLs of the backend etcd")
	flagEtcdPrefix  = flag.String("etcd-prefix", "", "etcd prefix")
	flagEtcdTimeout = flag.String("etcd-timeout", "2s", "dial timeout to etcd")
)

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

	s := &cmd.HTTPServer{
		Server: &http.Server{
			Addr:    *flagHTTP,
			Handler: r,
		},
		ShutdownTimeout: 3 * time.Minute,
	}
	s.ListenAndServe()

	err = cmd.Wait()
	if !cmd.IsSignaled(err) && err != nil {
		log.ErrorExit(err)
	}
}
