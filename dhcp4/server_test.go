package dhcp4

import (
	"bytes"
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
	pkt, err := createPacket(dhcp4.MsgOffer, nil)
	intfs, err := net.Interfaces()
	if err != nil {
		t.Fatal(err)
	}
	intf := intfs[0]

	dhcp := New("0.0.0.0:67", "lo", "", dhcp4Begin, dhcp4End).(*dhcpserver)
	err = dhcp.handleDiscover(&conn, pkt, &intf)
	if err != nil {
		t.Fatal(err)
	}

	if len(conn.packets) != 1 {
		t.Fatal("dhcp4.Server should return only one packet")
	}
	p := conn.packets[0]
	assertEqualPackets(t, expected, p)
}

func assertEqualPackets(t *testing.T, expected *dhcp4.Packet, actual *dhcp4.Packet) {
	if actual.Type != dhcp4.MsgOffer {
		t.Fatalf("Type expeceted: %d, actual: %d", actual.Type, dhcp4.MsgOffer)
	}
	if !expected.YourAddr.Equal(actual.YourAddr) {
		t.Fatalf("YourAddr expeceted: %v, actual: %v", expected.YourAddr, actual.YourAddr)
	}
	if !bytes.Equal(expected.Options[dhcp4.OptDHCPMessageType], actual.Options[dhcp4.OptDHCPMessageType]) {
		t.Fatalf("Options 53 expeceted: %v, actual: %v",
			expected.Options[dhcp4.OptDHCPMessageType], actual.Options[dhcp4.OptDHCPMessageType])
	}
}

func createPacket(msgType dhcp4.MessageType, options dhcp4.Options) (*dhcp4.Packet, error) {
	hwaddr, err := net.ParseMAC("00:00:00:00:00:00")
	if err != nil {
		return nil, err
	}
	pkt := dhcp4.Packet{
		Type:           msgType,
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
	pkt.Options[dhcp4.OptDHCPMessageType] = []byte{byte(msgType)}
	for k, v := range options {
		pkt.Options[k] = v
	}
	return &pkt, nil
}

func TestDHCP(t *testing.T) {
	t.Run("discover", testDiscover)
}
