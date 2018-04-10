package main

import (
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
)

type etcdConfig struct {
	Servers []string
	Prefix  string
}

func main() {
	flag.Parse()

	var e etcdConfig
	e.Servers = strings.Split(*flagEtcdServers, ",")
	e.Prefix = "/" + *flagEtcdPrefix

	cfg := clientv3.Config{
		Endpoints: e.Servers,
	}
	c, err := clientv3.New(cfg)
	if err != nil {
		log.ErrorExit(err)
	}

	r := mux.NewRouter()
	sabakan.InitConfig(r.PathPrefix("/api/v1/").Subrouter(), c, e.Prefix)
	sabakan.InitCrypts(r.PathPrefix("/api/v1/").Subrouter(), c, e.Prefix)

	s := &cmd.HTTPServer{
		Server: &http.Server{
			Addr:    *flagHTTP,
			Handler: r,
		},
		ShutdownTimeout: 3 * time.Minute,
	}
	s.ListenAndServe()

	err = cmd.Wait()
	if err != nil && !cmd.IsSignaled(err) {
		log.ErrorExit(err)
	}
}
