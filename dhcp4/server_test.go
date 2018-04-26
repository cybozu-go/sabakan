package dhcp4

import (
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
	c.packets = append(c.packets, pkt)
	return nil
}

var dhcp4Begin = net.IPv4(10, 69, 0, 33)
var dhcp4End = net.IPv4(10, 69, 0, 63)
var serverInterface net.Interface

func init() {
	intfs, _ := net.Interfaces()
	serverInterface = intfs[0]
}

func testDiscover(t *testing.T) {
	conn := dummyConn{}
	pkt := createPacket(dhcp4.MsgDiscover, nil)

	dhcp := New("0.0.0.0:67", "lo", "", dhcp4Begin, dhcp4End).(*dhcpserver)
	err := dhcp.handleDiscover(&conn, pkt, &serverInterface)
	if err != nil {
		t.Fatal(err)
	}

	if len(conn.packets) != 1 {
		t.Fatal("dhcp4.Server should return only one packet")
	}
	actual := conn.packets[0]
	expected := createPacket(dhcp4.MsgOffer, nil)
	expected.YourAddr = dhcp4Begin

	assertEqualPackets(t, expected, actual)
}

func testRequest(t *testing.T) {
	conn := dummyConn{}
	pkt := createPacket(dhcp4.MsgRequest, nil)

	dhcp := New("0.0.0.0:67", "lo", "", dhcp4Begin, dhcp4End).(*dhcpserver)
	err := dhcp.handleRequest(&conn, pkt, &serverInterface)
	if err != nil {
		t.Fatal(err)
	}

	if len(conn.packets) != 1 {
		t.Fatal("dhcp4.Server should return only one packet")
	}
	actual := conn.packets[0]
	expected := createPacket(dhcp4.MsgAck, nil)

	assertEqualPackets(t, expected, actual)
}

func assertEqualPackets(t *testing.T, expected *dhcp4.Packet, actual *dhcp4.Packet) {
	if actual.Type != expected.Type {
		t.Fatalf("Type expeceted: %d, actual: %d", expected.Type, actual.Type)
	}
	if !expected.YourAddr.Equal(actual.YourAddr) {
		t.Fatalf("YourAddr expeceted: %v, actual: %v", expected.YourAddr, actual.YourAddr)
	}
}

func createPacket(msgType dhcp4.MessageType, options dhcp4.Options) *dhcp4.Packet {
	hwaddr, _ := net.ParseMAC("00:00:00:00:00:00")
	pkt := &dhcp4.Packet{
		Type:           msgType,
		TransactionID:  []byte{1, 2, 3, 4},
		Broadcast:      true,
		HardwareAddr:   hwaddr,
		ClientAddr:     nil,
		YourAddr:       nil,
		ServerAddr:     nil,
		RelayAddr:      nil,
		BootServerName: "",
		BootFilename:   "",
		Options:        make(dhcp4.Options),
	}
	for k, v := range options {
		pkt.Options[k] = v
	}
	return pkt
}

func TestDHCP(t *testing.T) {
	t.Run("discover", testDiscover)
	t.Run("request", testRequest)
}
