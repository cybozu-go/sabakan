package dhcpd

import (
	"context"
	"net"

	"go.universe.tf/netboot/dhcp4"
)

func (h DHCPHandler) handleDiscover(ctx context.Context, pkt *dhcp4.Packet, intf *net.Interface) (*dhcp4.Packet, error) {
	serverAddr, err := getIPv4AddrForInterface(intf)
	if err != nil {
		return nil, err
	}

	ifaddr := pkt.RelayAddr
	if ifaddr == nil || ifaddr.IsUnspecified() {
		ifaddr = serverAddr
	}
	yourip, err := h.DHCP.Lease(ctx, ifaddr, pkt.HardwareAddr)
	if err != nil {
		return nil, err
	}
	secs, err := leaseSeconds(h.Model)
	if err != nil {
		return nil, err
	}
	resp := &dhcp4.Packet{
		Type:           dhcp4.MsgOffer,
		TransactionID:  pkt.TransactionID,
		Broadcast:      pkt.Broadcast,
		HardwareAddr:   pkt.HardwareAddr,
		YourAddr:       yourip,
		ServerAddr:     serverAddr,
		RelayAddr:      pkt.RelayAddr,
		BootServerName: serverAddr.String(),
		Options: dhcp4.Options{
			dhcp4.OptLeaseTime:        secs,
			dhcp4.OptServerIdentifier: serverAddr,
		},
	}
	return resp, nil
}
