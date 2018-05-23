package dhcpd

import (
	"bytes"
	"context"
	"net"
	"testing"

	"github.com/cybozu-go/sabakan"
	"github.com/cybozu-go/sabakan/models/mock"
	"go.universe.tf/netboot/dhcp4"
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

func testInterface() Interface {
	v6, v6net, _ := net.ParseCIDR("2001:db8::/32")
	v6net.IP = v6

	v4, v4net, _ := net.ParseCIDR("10.69.1.3/26")
	v4net.IP = v4

	return mockInterface{
		name:  "mock1",
		addrs: []net.Addr{v6net, v4net},
	}
}

func testIPEqual(t *testing.T, name string, respIP net.IP, expectedIP net.IP) {
	if expectedIP.Equal(net.IPv4zero) {
		if respIP == nil {
			return
		}
	}
	if !respIP.Equal(expectedIP) {
		t.Error(`wrong resp.`+name, respIP, expectedIP)
	}
}

func testComparePacket(t *testing.T, resp, expected *dhcp4.Packet) {
	if resp.Type != expected.Type {
		t.Error(`wrong resp.Type:`, resp.Type, expected.Type)
	}

	if !bytes.Equal(resp.TransactionID, expected.TransactionID) {
		t.Error(`wrong resp.TransactionID:`, resp.TransactionID, expected.TransactionID)
	}

	if resp.Broadcast != expected.Broadcast {
		t.Error(`wrong resp.Broadcast:`, resp.Broadcast, expected.Broadcast)
	}

	if !bytes.Equal(resp.HardwareAddr, expected.HardwareAddr) {
		t.Error(`wrong resp.HardwareAddr:`, resp.HardwareAddr, expected.HardwareAddr)
	}

	testIPEqual(t, "ClientAddr", resp.ClientAddr, expected.ClientAddr)
	testIPEqual(t, "YourAddr", resp.YourAddr, expected.YourAddr)
	testIPEqual(t, "ServerAddr", resp.ServerAddr, expected.ServerAddr)
	testIPEqual(t, "RelayAddr", resp.RelayAddr, expected.RelayAddr)

	if resp.BootServerName != expected.BootServerName {
		t.Error(`wrong resp.BootServerName:`, resp.BootServerName, expected.BootServerName)
	}

	if resp.BootFilename != expected.BootFilename {
		t.Error(`wrong resp.BootFilename:`, resp.BootFilename, expected.BootFilename)
	}

	for k, v := range expected.Options {
		v2, ok := resp.Options[k]
		if !ok {
			t.Error(`missing Option`, k)
			continue
		}
		if !bytes.Equal(v, v2) {
			t.Error(`wrong Option:`, k, v2, v)
		}
	}
}
