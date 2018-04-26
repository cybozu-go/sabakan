package dhcp4

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/cybozu-go/log"
	"go.universe.tf/netboot/dhcp4"
)

// Server is a DHCP server
type Server interface {
	Serve(ctx context.Context) error
}

// New creates a new dhcp Server object
func New(bind string, ifname string, ipxe string, begin, end net.IP) Server {
	s := new(dhcpserver)
	s.bind = bind
	s.ifname = ifname
	s.ipxe = ipxe
	s.assign = assignment{
		begin:  begin,
		end:    end,
		leases: make(map[uint32]struct{}),
	}
	return s
}

type dhcpserver struct {
	bind   string
	ifname string
	ipxe   string

	assign assignment
}

type dhcpConn interface {
	Close() error
	RecvDHCP() (*dhcp4.Packet, *net.Interface, error)
	SendDHCP(pkt *dhcp4.Packet, intf *net.Interface) error
}

func (s *dhcpserver) handleDiscover(conn dhcpConn, pkt *dhcp4.Packet, intf *net.Interface) error {
	ip, err := s.assign.next()
	if err != nil {
		log.Info("DHCP: Couldn't allocate ip address", map[string]interface{}{
			"mac_address": pkt.HardwareAddr,
			"error":       err,
		})
		return err
	}

	serverIP, err := interfaceIP(intf)
	if err != nil {
		log.Info("DHCP: Couldn't get a source address", map[string]interface{}{
			"mac_address": pkt.HardwareAddr,
			"interface":   intf.Name,
			"error":       err,
		})
		return err
	}

	resp, err := s.offerDHCP(pkt, serverIP, ip)
	if err != nil {
		log.Info("DHCP: Failed to construct DHCP Offer", map[string]interface{}{
			"mac_address": pkt.HardwareAddr,
			"error":       err,
		})
		return err
	}

	if err = conn.SendDHCP(resp, intf); err != nil {
		log.Info("DHCP: Failed to send DHCP Offer", map[string]interface{}{
			"mac_address": pkt.HardwareAddr,
			"error":       err,
		})
	}
	return err
}

func (s *dhcpserver) handleRequest(conn dhcpConn, pkt *dhcp4.Packet, intf *net.Interface) error {
	ip := pkt.Options[dhcp4.OptRequestedIP]

	serverIP, err := interfaceIP(intf)

	resp, err := s.ackDHCP(pkt, serverIP, ip)
	if err != nil {
		log.Info("DHCP: Failed to construct DHCP Ack", map[string]interface{}{
			"mac_address": pkt.HardwareAddr,
			"error":       err,
		})
		return err
	}

	if err = conn.SendDHCP(resp, intf); err != nil {
		log.Info("DHCP: Failed to send DHCP Ack", map[string]interface{}{
			"mac_address": pkt.HardwareAddr,
			"error":       err,
		})
	}
	return err
}

func (s *dhcpserver) Serve(ctx context.Context) error {
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
			log.Error("unknown packet type: %v",map[string]interface{}{
				"type": pkt.Type,
			})
		}
	}
}

func (s *dhcpserver) offerDHCP(pkt *dhcp4.Packet, serverIP net.IP, clientIP net.IP) (*dhcp4.Packet, error) {
	resp := &dhcp4.Packet{
		Type:          dhcp4.MsgOffer,
		TransactionID: pkt.TransactionID,
		Broadcast:     true,
		HardwareAddr:  pkt.HardwareAddr,
		RelayAddr:     pkt.RelayAddr,
		ServerAddr:    serverIP,
		YourAddr:      clientIP,
		Options:       make(dhcp4.Options),
	}

	return resp, nil
}

func (s *dhcpserver) ackDHCP(pkt *dhcp4.Packet, serverIP net.IP, clientIP net.IP) (*dhcp4.Packet, error) {
	resp := &dhcp4.Packet{
		Type:          dhcp4.MsgAck,
		TransactionID: pkt.TransactionID,
		Broadcast:     true,
		HardwareAddr:  pkt.HardwareAddr,
		RelayAddr:     pkt.RelayAddr,
		ServerAddr:    serverIP,
		YourAddr:      clientIP,
		Options:       make(dhcp4.Options),
	}

	return resp, nil
}

func interfaceIP(intf *net.Interface) (net.IP, error) {
	addrs, err := intf.Addrs()
	if err != nil {
		return nil, err
	}

	// Try to find an IPv4 address to use, in the following order:
	// global unicast (includes rfc1918), link-local unicast,
	// loopback.
	fs := [](func(net.IP) bool){
		net.IP.IsGlobalUnicast,
		net.IP.IsLinkLocalUnicast,
		net.IP.IsLoopback,
	}
	for _, f := range fs {
		for _, a := range addrs {
			ipaddr, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipaddr.IP.To4()
			if ip == nil {
				continue
			}
			if f(ip) {
				return ip, nil
			}
		}
	}

	return nil, errors.New("no usable unicast address configured on interface")
}
