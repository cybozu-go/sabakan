package main

import (
	"flag"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/log"
	"github.com/gorilla/mux"
)

var (
	flagHTTP           = flag.String("http", "0.0.0.0:8888", "<Listen IP>:<Port number>")
	flagEtcdServers    = flag.String("etcd-servers", "http://localhost:2379", "URLs of the backend etcd")
	flagEtcdPrefix     = flag.String("etcd-prefix", "", "etcd prefix")
	flagNodeIPv4Offset = flag.String("node-ipv4-offset", "", "IP address offset to assign Nodes")
	flagNodeRackShift  = flag.String("node-rack-shift", "", "Integer to calculate IP addresses for address each nodes based on --node-ipv4-offset")
	flagBMCIPv4Offset  = flag.String("bmc-ipv4-offset", "", "IP address offset to assign Baseboard Management Controller")
	flagBMCRackShift   = flag.String("bmc-rack-shift", "", "Integer to calculate IP addresses for address each BMC based on --bmc-ipv4-offset")
	flagNodeIPPerNode  = flag.String("node-ip-per-node", "1", "Number of IP addresses per node. Exclude BMC. Default to 1")
	flagBMCPerNode     = flag.String("bmc-ip-per-node", "1", "Number of IP addresses per BMC. Default to 1")
)

type etcdConfig struct {
	Servers []string
	Prefix  string
}

type etcdClient struct {
	c *clientv3.Client
}

func main() {
	flag.Parse()

	var e etcdConfig
	e.Servers = strings.Split(*flagEtcdServers, ",")
	e.Prefix = *flagEtcdPrefix

	cfg := clientv3.Config{
		Endpoints: e.Servers,
	}
	c, err := clientv3.New(cfg)
	if err != nil {
		log.ErrorExit(err)
	}

	r := mux.NewRouter()
	initHello(r.PathPrefix("/").Subrouter(), c)

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

func (e *etcdClient) initHelloFunc(r *mux.Router) {
	r.HandleFunc("/hello", e.handleHello).Methods("GET")
}

func initHello(r *mux.Router, c *clientv3.Client) {
	e := &etcdClient{c}
	e.initHelloFunc(r)
}

func (e *etcdClient) handleHello(w http.ResponseWriter, r *http.Request) {
	_, err := e.c.Put(r.Context(), "/world", "Hello, world v3")
	if err != nil {
		w.Write([]byte(err.Error() + "\n"))
		return
	}
	w.Write([]byte("Hello, world\n"))
}
