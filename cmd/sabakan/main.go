package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/log"
	"github.com/gorilla/mux"
)

var (
	flagHttp           = flag.String("http", "0.0.0.0:8888", "<Listen IP>:<Port number>")
	flagEtcdServers    = flag.String("etcd-servers", "", "URLs of the backend etcd")
	flagEtcdPrefix     = flag.String("etcd-prefix", "", "etcd prefix")
	flagNodeIPv4Offset = flag.String("node-ipv4-offset", "", "IP address offset to assign Nodes")
	flagNodeRackShift  = flag.String("node-rack-shift", "", "Integer to calculate IP addresses for address each nodes based on --node-ipv4-offset")
	flagBMCIPv4Offset  = flag.String("bmc-ipv4-offset", "", "IP address offset to assign Baseboard Management Controller")
	flagBMCRackShift   = flag.String("bmc-rack-shift", "", "Integer to calculate IP addresses for address each BMC based on --bmc-ipv4-offset")
	flagNodeIPPerNode  = flag.String("node-ip-per-node", "1", "Number of IP addresses per node. Exclude BMC. Default to 1")
	flagBMCPerNode     = flag.String("bmc-ip-per-node", "1", "Number of IP addresses per BMC. Default to 1")
)

type EtcdConfig struct {
	Servers []string
	Prefix  string
}

type EtcdClient struct {
	c client.Client
}

func main() {
	flag.Parse()

	var e EtcdConfig
	e.Servers = strings.Split(*flagEtcdServers, ",")
	e.Prefix = *flagEtcdPrefix

	cfg := client.Config{
		Endpoints: e.Servers,
		Transport: client.DefaultTransport,
	}
	c, err := client.New(cfg)
	if err != nil {
		// handle error
	}

	r := mux.NewRouter()
	initHello(r.PathPrefix("/").Subrouter(), c)

	s := &cmd.HTTPServer{
		Server: &http.Server{
			Addr:    fmt.Sprintf(":%d", 8080),
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

func (e *EtcdClient) initHelloFunc(r *mux.Router) {
	r.HandleFunc("/hello", e.handleHello).Methods("GET")
}

func initHello(r *mux.Router, c client.Client) {
	e := &EtcdClient{c}
	e.initHelloFunc(r)
}

func (e *EtcdClient) handleHello(w http.ResponseWriter, r *http.Request) {
	kAPI := client.NewKeysAPI(e.c)

	// create a new key /foo with the value "bar"
	_, err := kAPI.Create(r.Context(), "/hello", "Hello, world")
	if err != nil {
		// handle error
	}
	//r.Write([]byte("Hello, world"))
}
