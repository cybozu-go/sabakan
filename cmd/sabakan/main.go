package main

import (
	"github.com/gorilla/mux"
	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/log"
	"net/http"
	"fmt"
	"time"
	"flag"
)

var (
	flagHttp = flag.String("http", "0.0.0.0:8888", "<Listen IP>:<Port number>")
	flagEtcdServers = flag.String("etcd-servers", "", "URLs of the backend etcd")
	flagEtcdPrefix = flag.String("etcd-prefix", "", "etcd prefix")
	flagNodeIPv4Offset = flag.String("node-ipv4-offset", "", "IP address offset to assign Nodes")
	flagNodeRackShift = flag.String("node-rack-shift", "", "Integer to calculate IP addresses for address each nodes based on --node-ipv4-offset")
	flagBMCIPv4Offset = flag.String("bmc-ipv4-offset", "", "IP address offset to assign Baseboard Management Controller")
	flagBMCRackShift = flag.String("bmc-rack-shift", "", "Integer to calculate IP addresses for address each BMC based on --bmc-ipv4-offset")
	flagNodeIPPerNode = flag.String("node-ip-per-node", "1", "Number of IP addresses per node. Exclude BMC. Default to 1")
	flagBMCPerNode = flag.String("bmc-ip-per-node", "1", "Number of IP addresses per BMC. Default to 1")
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/hello", handleHello).Methods("GET")

	s := &cmd.HTTPServer{
		Server: &http.Server{
			Addr:    fmt.Sprintf(":%d", 8080),
			Handler: r,
		},
		ShutdownTimeout: 3 * time.Minute,
	}
	s.ListenAndServe()

	err := cmd.Wait()
	if err != nil && !cmd.IsSignaled(err) {
		log.ErrorExit(err)
	}
}

func handleHello(responseWriter http.ResponseWriter, request *http.Request) {
	responseWriter.Write([]byte ("Hello, world"))
}
