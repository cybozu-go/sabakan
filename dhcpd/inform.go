package dhcpd

import (
	"context"

	"go.universe.tf/netboot/dhcp4"
)

func (h DHCPHandler) handleInform(ctx context.Context, pkt *dhcp4.Packet, intf Interface) (*dhcp4.Packet, error) {
	serverAddr, err := getIPv4AddrForInterface(intf)
	if err != nil {
		return nil, err
	}

	opts := make(dhcp4.Options)
	opts[dhcp4.OptServerIdentifier] = serverAddr
	resp := &dhcp4.Packet{
		Type:           dhcp4.MsgAck,
		TransactionID:  pkt.TransactionID,
		Broadcast:      pkt.Broadcast,
		HardwareAddr:   pkt.HardwareAddr,
		ClientAddr:     pkt.ClientAddr,
		ServerAddr:     serverAddr,
		RelayAddr:      pkt.RelayAddr,
		BootServerName: serverAddr.String(),
		Options:        opts,
	}

	return resp, nil
}
