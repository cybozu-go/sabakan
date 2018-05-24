package dhcpd

import (
	"context"
	"net"
	"testing"

	"go.universe.tf/netboot/dhcp4"
)

func testReleasePacket() *dhcp4.Packet {
	txnID := []byte{0xaa, 0xbb, 0xcc, 0xdd}
	return &dhcp4.Packet{
		Type:          dhcp4.MsgRelease,
		TransactionID: txnID,
		Broadcast:     false,
		HardwareAddr:  []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
		ClientAddr:    []byte{10, 69, 1, 32},
		YourAddr:      net.IPv4zero,
		ServerAddr:    net.IPv4zero,
		RelayAddr:     net.IPv4zero,
		Options:       make(dhcp4.Options),
	}
}

func testRelease(t *testing.T) {
	t.Parallel()

	h := testNewHandler(26, 1, 0)

	pkt := testReleasePacket()
	pkt.Options[dhcp4.OptServerIdentifier] = []byte{10, 69, 1, 3}
	intf := testInterface()

	resp, err := h.handleRelease(context.Background(), pkt, intf)
	if err != errNoAction {
		t.Error("invalid error:", err)
	}
	if resp != nil {
		t.Error("unknown resp error")
	}
}

func TestRelease(t *testing.T) {
	t.Run("Release", testRelease)
}
