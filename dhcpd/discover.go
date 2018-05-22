package dhcpd

import (
	"context"
	"net"
	"time"

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
	} else {
		// To delay answer to relayed requests, sleep shortly.
		time.Sleep(50 * time.Millisecond)
	}

	yourip, err := h.DHCP.Lease(ctx, ifaddr, pkt.HardwareAddr)
	if err != nil {
		return nil, err
	}
	opts, err := h.makeOptions(yourip)
	if err != nil {
		return nil, err
	}
	opts[dhcp4.OptServerIdentifier] = serverAddr
	resp := &dhcp4.Packet{
		Type:           dhcp4.MsgOffer,
		TransactionID:  pkt.TransactionID,
		Broadcast:      pkt.Broadcast,
		HardwareAddr:   pkt.HardwareAddr,
		YourAddr:       yourip,
		ServerAddr:     serverAddr,
		RelayAddr:      pkt.RelayAddr,
		BootServerName: serverAddr.String(),
		Options:        opts,
	}
	return resp, nil
}
