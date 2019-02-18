package web

import (
	"context"
	"net"
	"net/url"
	"testing"

	"github.com/cybozu-go/sabakan/v2"
)

const testMyURL = "http://www.example.com"

func newTestServer(m sabakan.Model) *Server {
	// httptest.NewRequest() sets RemoteAddr as "192.0.2.1:1234"
	// https://golang.org/src/net/http/httptest/httptest.go?s=1162:1230#L31
	_, ipnet, _ := net.ParseCIDR("192.0.2.1/24")
	u, _ := url.Parse(testMyURL)
	return NewServer(m, "", u, []*net.IPNet{ipnet}, false)
}

func testWithIPAM(t *testing.T, m sabakan.Model) *sabakan.IPAMConfig {
	config := &sabakan.IPAMConfig{
		MaxNodesInRack:    28,
		NodeIPv4Pool:      "10.69.0.0/20",
		NodeIPv4Offset:    "",
		NodeRangeSize:     6,
		NodeRangeMask:     26,
		NodeIPPerNode:     3,
		NodeIndexOffset:   3,
		NodeGatewayOffset: 1,
		BMCIPv4Pool:       "10.72.16.0/20",
		BMCIPv4Offset:     "0.0.1.0",
		BMCRangeSize:      5,
		BMCRangeMask:      20,
		BMCGatewayOffset:  1,
	}
	err := m.IPAM.PutConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	return config
}
