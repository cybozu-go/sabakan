package dhcpd

import (
	"context"
	"encoding/binary"
	"net"
	"testing"

	"go.universe.tf/netboot/dhcp4"
)

func testDiscoverPacket() *dhcp4.Packet {
	txnID := []byte{0xaa, 0xbb, 0xcc, 0xdd}
	return &dhcp4.Packet{
		Type:          dhcp4.MsgDiscover,
		TransactionID: txnID,
		Broadcast:     true,
		HardwareAddr:  []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
		ClientAddr:    net.IPv4zero,
		YourAddr:      net.IPv4zero,
		ServerAddr:    net.IPv4zero,
		RelayAddr:     net.IPv4zero,
		Options:       make(dhcp4.Options),
	}
}

func testDiscoverInterface() Interface {
	v6, v6net, _ := net.ParseCIDR("2001:db8::/32")
	v6net.IP = v6

	v4, v4net, _ := net.ParseCIDR("10.69.1.3/26")
	v4net.IP = v4

	return mockInterface{
		name:  "mock1",
		addrs: []net.Addr{v6net, v4net},
	}
}

func testComparePacket(t *testing.T, resp, expected *dhcp4.Packet) {
	if resp.Type != expected.Type {
		t.Error(`wrong resp.Type:`, resp.Type, expected.Type)
	}
}

func testDiscoverDirect(t *testing.T) {
	t.Parallel()

	h := testNewHandler(26, 1, 0)

	pkt := testDiscoverPacket()
	intf := testDiscoverInterface()
	expected := testDiscoverPacket()
	expected.Type = dhcp4.MsgOffer
	expected.YourAddr = net.IPv4(10, 69, 1, 32)
	expected.ServerAddr = net.IPv4(10, 69, 1, 3)
	expected.BootServerName = "10.69.1.3"
	expected.Options[dhcp4.OptSubnetMask] = net.IPv4(255, 255, 255, 192)
	expected.Options[dhcp4.OptRouters] = net.IPv4(10, 69, 1, 1)
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, 3600)
	expected.Options[dhcp4.OptLeaseTime] = buf
	expected.Options[dhcp4.OptServerIdentifier] = net.IPv4(10, 69, 1, 3)

	resp, err := h.handleDiscover(context.Background(), pkt, intf)
	if err != nil {
		t.Fatal(err)
	}
	testComparePacket(t, resp, expected)
}

func testDiscoverRelayed(t *testing.T) {
	t.Parallel()
}

func testDiscoverHTTPBoot(t *testing.T) {
	t.Parallel()
}

func testDiscoverIPXE(t *testing.T) {
	t.Parallel()
}

func TestDiscover(t *testing.T) {
	t.Run("Direct", testDiscoverDirect)
	t.Run("Relayed", testDiscoverRelayed)
	t.Run("HTTPBoot", testDiscoverHTTPBoot)
	t.Run("iPXE", testDiscoverIPXE)
}
