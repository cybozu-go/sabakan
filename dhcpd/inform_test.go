package dhcpd

import (
	"context"
	"net"
	"testing"

	"go.universe.tf/netboot/dhcp4"
)

func testInformPacket() *dhcp4.Packet {
	txnID := []byte{0xaa, 0xbb, 0xcc, 0xdd}
	return &dhcp4.Packet{
		Type:          dhcp4.MsgInform,
		TransactionID: txnID,
		Broadcast:     true,
		HardwareAddr:  []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
		ClientAddr:    []byte{10, 69, 0, 196},
		YourAddr:      net.IPv4zero,
		ServerAddr:    net.IPv4zero,
		RelayAddr:     net.IPv4zero,
		Options:       make(dhcp4.Options),
	}
}

func testInformDirect(t *testing.T) {
	t.Parallel()

	h := testNewHandler(26, 1, 0)

	pkt := testInformPacket()
	intf := testInterface()
	expected := testInformPacket()
	expected.Type = dhcp4.MsgAck
	expected.ServerAddr = []byte{10, 69, 1, 3}
	expected.BootServerName = "10.69.1.3"
	expected.Options[dhcp4.OptServerIdentifier] = []byte{10, 69, 1, 3}

	resp, err := h.handleInform(context.Background(), pkt, intf)
	if err != nil {
		t.Fatal(err)
	}
	testComparePacket(t, resp, expected)
}

func testInformRelayed(t *testing.T) {
	t.Parallel()

	h := testNewHandler(26, 1, 0)

	pkt := testInformPacket()
	pkt.RelayAddr = []byte{10, 69, 0, 129}
	intf := testInterface()
	expected := testInformPacket()
	expected.Type = dhcp4.MsgAck
	expected.ServerAddr = []byte{10, 69, 1, 3}
	expected.RelayAddr = []byte{10, 69, 0, 129}
	expected.BootServerName = "10.69.1.3"
	expected.Options[dhcp4.OptServerIdentifier] = []byte{10, 69, 1, 3}

	resp, err := h.handleInform(context.Background(), pkt, intf)
	if err != nil {
		t.Fatal(err)
	}
	testComparePacket(t, resp, expected)
}

func TestInform(t *testing.T) {
	t.Run("Direct", testInformDirect)
	t.Run("Relayed", testInformRelayed)
}
