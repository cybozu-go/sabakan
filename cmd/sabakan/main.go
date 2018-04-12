package main

import (
	"context"
	"flag"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
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
	defer c.Close()

	sabakan.Indexing(c, e.Prefix)

	r := mux.NewRouter()
	etcdClient := &sabakan.EtcdClient{Client: c, Prefix: e.Prefix}
	sabakan.InitConfig(r.PathPrefix("/api/v1/").Subrouter(), etcdClient)
	sabakan.InitCrypts(r.PathPrefix("/api/v1/").Subrouter(), etcdClient)
	sabakan.InitMachines(r.PathPrefix("/api/v1/").Subrouter(), etcdClient)

	// Monitor changes to keys and update index
	go func() {
		cfg := clientv3.Config{
			Endpoints: e.Servers,
		}
		c, err := clientv3.New(cfg)
		if err != nil {
			log.ErrorExit(err)
		}
		defer c.Close()

		key := path.Join(e.Prefix, sabakan.EtcdKeyMachines)
		rch := c.Watch(context.TODO(), key, clientv3.WithPrefix(), clientv3.WithPrevKV())
		for wresp := range rch {
			for _, ev := range wresp.Events {
				if ev.Type == mvccpb.PUT && ev.PrevKv != nil {
					sabakan.UpdateIndex(ev.PrevKv.Value, ev.Kv.Value)
				}
				if ev.Type == mvccpb.PUT && ev.PrevKv == nil {
					sabakan.AddIndex(ev.Kv.Value)
				}
				if ev.Type == mvccpb.DELETE {
					sabakan.DeleteIndex(ev.PrevKv.Value)
				}
			}
		}
	}()

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
