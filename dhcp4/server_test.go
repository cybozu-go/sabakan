package dhcp4

import (
	"fmt"
	"net"
	"testing"

	"go.universe.tf/netboot/dhcp4"
)

type dummyConn struct {
	packets []*dhcp4.Packet
}

func (c *dummyConn) Close() error {
	return nil
}

func (c *dummyConn) RecvDHCP() (*dhcp4.Packet, *net.Interface, error) {
	return nil, nil, nil
}

func (c *dummyConn) SendDHCP(pkt *dhcp4.Packet, intf *net.Interface) error {
	fmt.Printf("Send: %v\n", pkt)
	c.packets = append(c.packets, pkt)
	return nil
}

func testDiscover(t *testing.T) {
	dhcp4Begin := net.IPv4(10, 69, 0, 33)
	dhcp4End := net.IPv4(10, 69, 0, 63)

	conn := dummyConn{}
	hwaddr, err := net.ParseMAC("00:00:00:00:00:00")
	if err != nil {
		t.Fatal(err)
	}
	pkt := dhcp4.Packet{
		Type:           dhcp4.MsgDiscover,
		TransactionID:  []byte{1, 2, 3, 4},
		Broadcast:      true,
		HardwareAddr:   hwaddr,
		ClientAddr:     net.ParseIP("0.0.0.0"),
		YourAddr:       net.ParseIP("0.0.0.0"),
		ServerAddr:     net.ParseIP("0.0.0.0"),
		RelayAddr:      net.ParseIP("0.0.0.0"),
		BootServerName: "",
		BootFilename:   "",
		Options:        make(dhcp4.Options),
	}
	pkt.Options[dhcp4.OptDHCPMessageType] = []byte{1}
	intf := net.Interface{}

	dhcp := New("0.0.0.0:67", "", "", dhcp4Begin, dhcp4End).(*dhcpserver)
	err = dhcp.handleDiscover(&conn, &pkt, &intf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDHCP(t *testing.T) {
	t.Run("discover", testDiscover)
}
