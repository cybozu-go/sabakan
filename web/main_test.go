package web

import (
	"net"

	"github.com/cybozu-go/sabakan"
)

func newTestServer(m sabakan.Model) *Server {
	_, ipnet, err := net.ParseCIDR("192.0.2.1/32")
	if err != nil {
		panic(err)
	}
	return &Server{
		Model:         m,
		AllowdRemotes: []*net.IPNet{ipnet},
	}
}
