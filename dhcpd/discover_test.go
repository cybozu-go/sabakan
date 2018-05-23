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

func testDiscoverDirect(t *testing.T) {
	t.Parallel()

	h := testNewHandler(26, 1, 0)

	pkt := testDiscoverPacket()
	intf := testInterface()
	expected := testDiscoverPacket()
	expected.Type = dhcp4.MsgOffer
	expected.YourAddr = []byte{10, 69, 1, 32}
	expected.ServerAddr = []byte{10, 69, 1, 3}
	expected.BootServerName = "10.69.1.3"
	expected.Options[dhcp4.OptSubnetMask] = []byte{255, 255, 255, 192}
	expected.Options[dhcp4.OptRouters] = []byte{10, 69, 1, 1}
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, 3600)
	expected.Options[dhcp4.OptLeaseTime] = buf
	expected.Options[dhcp4.OptServerIdentifier] = []byte{10, 69, 1, 3}

	resp, err := h.handleDiscover(context.Background(), pkt, intf)
	if err != nil {
		t.Fatal(err)
	}
	testComparePacket(t, resp, expected)

	h = testNewHandler(24, 100, 10)
	expected.Options[dhcp4.OptSubnetMask] = []byte{255, 255, 255, 0}
	expected.Options[dhcp4.OptRouters] = []byte{10, 69, 1, 100}
	binary.BigEndian.PutUint32(buf, 600)
	expected.Options[dhcp4.OptLeaseTime] = buf

	resp, err = h.handleDiscover(context.Background(), pkt, intf)
	if err != nil {
		t.Fatal(err)
	}
	testComparePacket(t, resp, expected)
}

func testDiscoverRelayed(t *testing.T) {
	t.Parallel()

	h := testNewHandler(26, 1, 0)

	pkt := testDiscoverPacket()
	pkt.RelayAddr = []byte{10, 69, 0, 129}
	intf := testInterface()
	expected := testDiscoverPacket()
	expected.Type = dhcp4.MsgOffer
	expected.YourAddr = []byte{10, 69, 0, 160}
	expected.ServerAddr = []byte{10, 69, 1, 3}
	expected.RelayAddr = []byte{10, 69, 0, 129}
	expected.BootServerName = "10.69.1.3"
	expected.Options[dhcp4.OptSubnetMask] = []byte{255, 255, 255, 192}
	expected.Options[dhcp4.OptRouters] = []byte{10, 69, 0, 129}
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, 3600)
	expected.Options[dhcp4.OptLeaseTime] = buf
	expected.Options[dhcp4.OptServerIdentifier] = []byte{10, 69, 1, 3}

	resp, err := h.handleDiscover(context.Background(), pkt, intf)
	if err != nil {
		t.Fatal(err)
	}
	testComparePacket(t, resp, expected)
}

func testDiscoverHTTPBoot(t *testing.T) {
	t.Parallel()

	h := testNewHandler(26, 1, 0)

	pkt := testDiscoverPacket()
	pkt.Options[93] = []byte{0x00, 0x10}
	pkt.Options[dhcp4.OptVendorIdentifier] = []byte("HTTPClient:hogehoge")
	intf := testInterface()
	expected := testDiscoverPacket()
	expected.Type = dhcp4.MsgOffer
	expected.YourAddr = []byte{10, 69, 1, 32}
	expected.ServerAddr = []byte{10, 69, 1, 3}
	expected.BootServerName = "10.69.1.3"
	expected.Options[dhcp4.OptSubnetMask] = []byte{255, 255, 255, 192}
	expected.Options[dhcp4.OptRouters] = []byte{10, 69, 1, 1}
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, 3600)
	expected.Options[dhcp4.OptLeaseTime] = buf
	expected.Options[dhcp4.OptServerIdentifier] = []byte{10, 69, 1, 3}
	expected.Options[dhcp4.OptVendorIdentifier] = []byte("HTTPClient")
	expected.BootFilename = "http://10.69.1.3:80/api/v1/boot/ipxe.efi"

	resp, err := h.handleDiscover(context.Background(), pkt, intf)
	if err != nil {
		t.Fatal(err)
	}
	testComparePacket(t, resp, expected)
}

func testDiscoverIPXE(t *testing.T) {
	t.Parallel()

	h := testNewHandler(26, 1, 0)

	pkt := testDiscoverPacket()
	pkt.Options[77] = []byte("iPXE")
	intf := testInterface()
	expected := testDiscoverPacket()
	expected.Type = dhcp4.MsgOffer
	expected.YourAddr = []byte{10, 69, 1, 32}
	expected.ServerAddr = []byte{10, 69, 1, 3}
	expected.BootServerName = "10.69.1.3"
	expected.Options[dhcp4.OptSubnetMask] = []byte{255, 255, 255, 192}
	expected.Options[dhcp4.OptRouters] = []byte{10, 69, 1, 1}
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, 3600)
	expected.Options[dhcp4.OptLeaseTime] = buf
	expected.Options[dhcp4.OptServerIdentifier] = []byte{10, 69, 1, 3}
	expected.BootFilename = "http://10.69.1.3:80/api/v1/boot/coreos/ipxe"

	resp, err := h.handleDiscover(context.Background(), pkt, intf)
	if err != nil {
		t.Fatal(err)
	}
	testComparePacket(t, resp, expected)

}

func TestDiscover(t *testing.T) {
	t.Run("Direct", testDiscoverDirect)
	t.Run("Relayed", testDiscoverRelayed)
	t.Run("HTTPBoot", testDiscoverHTTPBoot)
	t.Run("iPXE", testDiscoverIPXE)
}
