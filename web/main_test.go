package web

import (
	"net"
	"net/url"

	"github.com/cybozu-go/sabakan"
)

func newTestServer(m sabakan.Model) *Server {
	return newTestServerWithURL(m, nil)
}

func newTestServerWithURL(m sabakan.Model, u *url.URL) *Server {
	// httptest.NewRequest() sets RemoteAddr as "192.0.2.1:1234"
	// https://golang.org/src/net/http/httptest/httptest.go?s=1162:1230#L31
	_, ipnet, err := net.ParseCIDR("192.0.2.1/24")
	if err != nil {
		panic(err)
	}
	return &Server{
		Model:          m,
		AllowedRemotes: []*net.IPNet{ipnet},
		MyURL:          u,
	}
}
