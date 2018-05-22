package dhcpd

import (
	"context"
	"errors"
	"net"

	"go.universe.tf/netboot/dhcp4"
)

func getIPv4AddrForInterface(intf *net.Interface) (net.IP, error) {
	addrs, err := intf.Addrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		ipaddr, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		if ipaddr.IP.To4() != nil {
			return ipaddr.IP, nil
		}
	}
	return nil, errors.New("No IPv4 address for " + intf.Name)
}

func (h DHCPHandler) handleDiscover(ctx context.Context, pkt *dhcp4.Packet, intf *net.Interface) (*dhcp4.Packet, error) {
	ifaddr := pkt.RelayAddr
	serverAddr, err := getIPv4AddrForInterface(intf)
	if err != nil {
		return nil, err
	}
	if ifaddr == nil || ifaddr.IsUnspecified() {
		ifaddr = serverAddr
	}
	yourip, err := h.Model.DHCP.Lease(ctx, ifaddr, pkt.HardwareAddr)
	if err != nil {
		return nil, err
	}
	resp := &dhcp4.Packet{
		Type:          dhcp4.MsgOffer,
		TransactionID: pkt.TransactionID,
		Broadcast:     pkt.Broadcast,
		HardwareAddr:  pkt.HardwareAddr,
		YourAddr:      yourip,
		ServerAddr:    serverAddr,
		RelayAddr:     pkt.RelayAddr,
		Options:       make(dhcp4.Options),
	}
	return resp, nil
}
