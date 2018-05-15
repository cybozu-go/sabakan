package dhcp4

import (
	"context"
	"fmt"
	"net"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan"
	"go.universe.tf/netboot/dhcp4"
)

// Server is a DHCP server
type dhcpServer struct {
	bind   string
	ifname string
	ipxe   string
	lessor sabakan.Lessor
}

// New creates a new dhcp Server object
func New(bind string, ifname string, ipxe string, lessor sabakan.Lessor) *dhcpServer {
	return &dhcpServer{
		bind:   bind,
		ifname: ifname,
		ipxe:   ipxe,
		lessor: lessor,
	}
}

func (s *dhcpServer) Serve(ctx context.Context) error {
	conn, err := dhcp4.NewConn(s.bind)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	for {
		pkt, intf, err := conn.RecvDHCP()
		if err != nil {
			return fmt.Errorf("Receiving DHCP packet: %s", err)
		}
		if intf.Name != s.ifname {
			log.Debug("DHCP: Ignoring packet", map[string]interface{}{
				"listen_interface": s.ifname,
				"received_on":      intf.Name,
			})
			continue
		}

		switch pkt.Type {
		case dhcp4.MsgDiscover:
			_ = s.handleDiscover(conn, pkt, intf)
		case dhcp4.MsgRequest:
			_ = s.handleRequest(conn, pkt, intf)
		default:
			log.Error("unknown packet type: %v", map[string]interface{}{
				"type": pkt.Type,
			})
		}
	}
}

func (s *dhcpServer) handleDiscover(conn *dhcp4.Conn, pkt *dhcp4.Packet, intf *net.Interface) error {
	resp, err := s.offer(pkt, intf)
	if err != nil {
		return err
	}
	err = conn.SendDHCP(resp, intf)
	return err
}

func (s *dhcpServer) handleRequest(conn *dhcp4.Conn, pkt *dhcp4.Packet, intf *net.Interface) error {
	resp, err := s.acknowledge(pkt, intf)
	if err != nil {
		return err
	}

	err = conn.SendDHCP(resp, intf)
	return err
}
