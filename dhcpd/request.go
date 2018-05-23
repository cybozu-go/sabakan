package dhcpd

import (
	"context"

	"go.universe.tf/netboot/dhcp4"
)

// DHCPREQUEST has three use-cases.
//   1. accept offer from a server.
//   2. confirm previously assigned IP address.
//   3. renew/rebind lease for an already received address.
//
// To distinguish these three, "server identifier" option (54) and
// "requested IP address" option (50) are used.
func (h DHCPHandler) handleRequest(ctx context.Context, pkt *dhcp4.Packet, intf Interface) (*dhcp4.Packet, error) {
	serverAddr, err := getIPv4AddrForInterface(intf)
	if err != nil {
		return nil, err
	}

	serverIdentifier, err := pkt.Options.IP(dhcp4.OptServerIdentifier)
	hasServerIdentifier := err == nil

	requestedIP, err := pkt.Options.IP(dhcp4.OptRequestedIP)
	hasRequestedIP := err == nil

	if hasServerIdentifier {
		// case 1.
		if !serverAddr.Equal(serverIdentifier) {
			// not chosen
			return nil, errNotChosen
		}

		resp, err := h.handleDiscover(ctx, pkt, intf)
		if err != nil {
			return nil, err
		}
		resp.Type = dhcp4.MsgAck
		return resp, nil
	}

	if hasRequestedIP {
		// case 2.
		err = h.DHCP.Renew(ctx, requestedIP, pkt.HardwareAddr)
		if err != nil {
			return nil, errNoRecord
		}

		opts, err := h.makeOptions(requestedIP)
		if err != nil {
			return nil, err
		}
		opts[dhcp4.OptServerIdentifier] = serverAddr
		resp := &dhcp4.Packet{
			Type:           dhcp4.MsgAck,
			TransactionID:  pkt.TransactionID,
			Broadcast:      pkt.Broadcast,
			HardwareAddr:   pkt.HardwareAddr,
			YourAddr:       requestedIP,
			ServerAddr:     serverAddr,
			RelayAddr:      pkt.RelayAddr,
			BootServerName: serverAddr.String(),
			Options:        opts,
		}

		return resp, nil
	}

	// case 3.
	err = h.DHCP.Renew(ctx, pkt.ClientAddr, pkt.HardwareAddr)
	if err != nil {
		return nil, errNoRecord
	}

	opts, err := h.makeOptions(pkt.ClientAddr)
	if err != nil {
		return nil, err
	}
	opts[dhcp4.OptServerIdentifier] = serverAddr
	resp := &dhcp4.Packet{
		Type:           dhcp4.MsgAck,
		TransactionID:  pkt.TransactionID,
		Broadcast:      pkt.Broadcast,
		HardwareAddr:   pkt.HardwareAddr,
		ClientAddr:     pkt.ClientAddr,
		YourAddr:       pkt.ClientAddr,
		ServerAddr:     serverAddr,
		RelayAddr:      pkt.RelayAddr,
		BootServerName: serverAddr.String(),
		Options:        opts,
	}
	return resp, nil
}