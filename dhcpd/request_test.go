package dhcpd

import (
	"context"
	"encoding/binary"
	"net"
	"testing"

	"go.universe.tf/netboot/dhcp4"
)

func testRequestPacket() *dhcp4.Packet {
	txnID := []byte{0xaa, 0xbb, 0xcc, 0xdd}
	return &dhcp4.Packet{
		Type:          dhcp4.MsgRequest,
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

func testRequestSelected(t *testing.T) {
	t.Parallel()

	h := testNewHandler(26, 1, 0)

	pkt := testRequestPacket()
	pkt.Options[dhcp4.OptServerIdentifier] = []byte{10, 69, 1, 3}
	intf := testInterface()
	expected := testRequestPacket()
	expected.Type = dhcp4.MsgAck
	expected.YourAddr = []byte{10, 69, 1, 32}
	expected.ServerAddr = []byte{10, 69, 1, 3}
	expected.BootServerName = "10.69.1.3"
	expected.Options[dhcp4.OptSubnetMask] = []byte{255, 255, 255, 192}
	expected.Options[dhcp4.OptRouters] = []byte{10, 69, 1, 1}
	expected.Options[dhcp4.OptDNSServers] = []byte{10, 0, 0, 1, 10, 0, 0, 2}
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, 3600)
	expected.Options[dhcp4.OptLeaseTime] = buf
	expected.Options[dhcp4.OptServerIdentifier] = []byte{10, 69, 1, 3}

	resp, err := h.handleRequest(context.Background(), pkt, intf)
	if err != nil {
		t.Fatal(err)
	}
	testComparePacket(t, resp, expected)
}

func testRequestNotSelected(t *testing.T) {
	t.Parallel()

	h := testNewHandler(26, 1, 0)

	pkt := testRequestPacket()
	pkt.Options[dhcp4.OptServerIdentifier] = []byte{10, 69, 1, 2}
	intf := testInterface()

	_, err := h.handleRequest(context.Background(), pkt, intf)
	if err != errNotChosen {
		t.Error("invalid error:", err)
	}
}

func testRequestRelayedSelected(t *testing.T) {
	t.Parallel()

	h := testNewHandler(26, 1, 0)

	pkt := testRequestPacket()
	pkt.RelayAddr = []byte{10, 69, 0, 129}
	pkt.Options[dhcp4.OptServerIdentifier] = []byte{10, 69, 1, 3}
	intf := testInterface()
	expected := testRequestPacket()
	expected.Type = dhcp4.MsgAck
	expected.YourAddr = []byte{10, 69, 0, 160}
	expected.ServerAddr = []byte{10, 69, 1, 3}
	expected.RelayAddr = []byte{10, 69, 0, 129}
	expected.BootServerName = "10.69.1.3"
	expected.Options[dhcp4.OptSubnetMask] = []byte{255, 255, 255, 192}
	expected.Options[dhcp4.OptRouters] = []byte{10, 69, 0, 129}
	expected.Options[dhcp4.OptDNSServers] = []byte{10, 0, 0, 1, 10, 0, 0, 2}
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, 3600)
	expected.Options[dhcp4.OptLeaseTime] = buf
	expected.Options[dhcp4.OptServerIdentifier] = []byte{10, 69, 1, 3}

	resp, err := h.handleRequest(context.Background(), pkt, intf)
	if err != nil {
		t.Fatal(err)
	}
	testComparePacket(t, resp, expected)
}

func testRequestConfirm(t *testing.T) {
	t.Parallel()

	h := testNewHandler(26, 1, 0)

	// Request before Discover; ignored
	pkt := testRequestPacket()
	pkt.Options[dhcp4.OptRequestedIP] = []byte{10, 69, 1, 33}
	intf := testInterface()

	_, err := h.handleRequest(context.Background(), pkt, intf)
	if err != errNoRecord {
		t.Error("invalid error:", err)
	}

	pkt = testDiscoverPacket()
	resp, err := h.handleDiscover(context.Background(), pkt, intf)
	if err != nil {
		t.Fatal(err)
	}

	pkt = testRequestPacket()
	pkt.Options[dhcp4.OptRequestedIP] = resp.YourAddr
	expected := testRequestPacket()
	expected.Type = dhcp4.MsgAck
	expected.YourAddr = resp.YourAddr
	expected.ServerAddr = []byte{10, 69, 1, 3}
	expected.BootServerName = "10.69.1.3"
	expected.Options[dhcp4.OptSubnetMask] = []byte{255, 255, 255, 192}
	expected.Options[dhcp4.OptRouters] = []byte{10, 69, 1, 1}
	expected.Options[dhcp4.OptDNSServers] = []byte{10, 0, 0, 1, 10, 0, 0, 2}
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, 3600)
	expected.Options[dhcp4.OptLeaseTime] = buf
	expected.Options[dhcp4.OptServerIdentifier] = []byte{10, 69, 1, 3}

	resp2, err := h.handleRequest(context.Background(), pkt, intf)
	if err != nil {
		t.Fatal(err)
	}
	testComparePacket(t, resp2, expected)
}

func testRequestRenew(t *testing.T) {
	t.Parallel()

	h := testNewHandler(26, 1, 0)

	// Request before Discover; ignored
	pkt := testRequestPacket()
	pkt.ClientAddr = []byte{10, 69, 1, 33}
	intf := testInterface()

	_, err := h.handleRequest(context.Background(), pkt, intf)
	if err != errNoRecord {
		t.Error("invalid error:", err)
	}

	pkt = testDiscoverPacket()
	resp, err := h.handleDiscover(context.Background(), pkt, intf)
	if err != nil {
		t.Fatal(err)
	}

	pkt = testRequestPacket()
	pkt.ClientAddr = resp.YourAddr
	expected := testRequestPacket()
	expected.Type = dhcp4.MsgAck
	expected.ClientAddr = resp.YourAddr
	expected.YourAddr = resp.YourAddr
	expected.ServerAddr = []byte{10, 69, 1, 3}
	expected.BootServerName = "10.69.1.3"
	expected.Options[dhcp4.OptSubnetMask] = []byte{255, 255, 255, 192}
	expected.Options[dhcp4.OptRouters] = []byte{10, 69, 1, 1}
	expected.Options[dhcp4.OptDNSServers] = []byte{10, 0, 0, 1, 10, 0, 0, 2}
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, 3600)
	expected.Options[dhcp4.OptLeaseTime] = buf
	expected.Options[dhcp4.OptServerIdentifier] = []byte{10, 69, 1, 3}

	resp2, err := h.handleRequest(context.Background(), pkt, intf)
	if err != nil {
		t.Fatal(err)
	}
	testComparePacket(t, resp2, expected)
}

func TestRequest(t *testing.T) {
	t.Run("Selected", testRequestSelected)
	t.Run("NotSelected", testRequestNotSelected)
	t.Run("RelayedSelected", testRequestRelayedSelected)
	t.Run("Confirm", testRequestConfirm)
	t.Run("Renew", testRequestRenew)
}
