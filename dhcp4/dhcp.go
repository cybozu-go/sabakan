package dhcp4

import (
	"errors"
	"net"

	"github.com/cybozu-go/log"
	"go.universe.tf/netboot/dhcp4"
)

func (s *DHCPServer) offer(pkt *dhcp4.Packet, intf *net.Interface) (*dhcp4.Packet, error) {

	ip, err := s.leaser.Lease()
	if err != nil {
		log.Info("DHCP: Couldn't allocate ip address", map[string]interface{}{
			"mac_address": pkt.HardwareAddr,
			"error":       err,
		})
		return nil, err
	}

	serverIP, err := s.interfaceIP(intf)
	if err != nil {
		return nil, err
	}

	resp := &dhcp4.Packet{
		Type:          dhcp4.MsgOffer,
		TransactionID: pkt.TransactionID,
		Broadcast:     true,
		HardwareAddr:  pkt.HardwareAddr,
		RelayAddr:     pkt.RelayAddr,
		ServerAddr:    serverIP,
		YourAddr:      ip,
		Options:       make(dhcp4.Options),
	}

	return resp, nil
}

func (s *DHCPServer) acknowledge(pkt *dhcp4.Packet, intf *net.Interface) (*dhcp4.Packet, error) {
	ip := pkt.Options[dhcp4.OptRequestedIP]

	serverIP, err := s.interfaceIP(intf)
	if err != nil {
		return nil, err
	}

	resp := &dhcp4.Packet{
		Type:          dhcp4.MsgAck,
		TransactionID: pkt.TransactionID,
		Broadcast:     true,
		HardwareAddr:  pkt.HardwareAddr,
		RelayAddr:     pkt.RelayAddr,
		ServerAddr:    serverIP,
		YourAddr:      ip,
		Options:       make(dhcp4.Options),
	}

	return resp, nil
}

func (s *DHCPServer) interfaceIP(intf *net.Interface) (net.IP, error) {
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
