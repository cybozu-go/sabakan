package dhcpd

import (
	"context"
	"net"

	"github.com/cybozu-go/sabakan"
	"github.com/cybozu-go/sabakan/models/mock"
)

type mockInterface struct {
	name  string
	addrs []net.Addr
}

func (i mockInterface) Addrs() ([]net.Addr, error) {
	return i.addrs, nil
}

func (i mockInterface) Name() string {
	return i.name
}

func testNewHandler(maskbits, gwoffset, leasemin uint) DHCPHandler {
	m := mock.NewModel()
	m.IPAM.PutConfig(context.Background(), &sabakan.IPAMConfig{
		MaxNodesInRack:  28,
		NodeIPv4Pool:    "10.69.0.0/20",
		NodeRangeSize:   6,
		NodeRangeMask:   maskbits,
		NodeIndexOffset: 3,
		NodeIPPerNode:   3,
		BMCIPv4Pool:     "10.72.16.0/20",
		BMCRangeSize:    5,
		BMCRangeMask:    20,
	})
	m.DHCP.PutConfig(context.Background(), &sabakan.DHCPConfig{
		GatewayOffset: gwoffset,
		LeaseMinutes:  leasemin,
	})

	return DHCPHandler{Model: m, URLPort: "80"}
}
