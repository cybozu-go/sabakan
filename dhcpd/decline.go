package dhcpd

import (
	"context"

	"github.com/cybozu-go/log"
	"go.universe.tf/netboot/dhcp4"
)

func (h DHCPHandler) handleDecline(ctx context.Context, pkt *dhcp4.Packet, intf Interface) (*dhcp4.Packet, error) {
	serverAddr, err := getIPv4AddrForInterface(intf)
	if err != nil {
		return nil, err
	}

	serverIdentifier, err := pkt.Options.IP(dhcp4.OptServerIdentifier)
	if err != nil {
		return nil, err
	}

	if !serverAddr.Equal(serverIdentifier) {
		log.Info("dhcp: ignored decline to another server", addPacketLog(pkt, map[string]interface{}{
			"serverid": serverIdentifier,
		}))
		return nil, errNotChosen
	}

	requestedIP, err := pkt.Options.IP(dhcp4.OptRequestedIP)
	if err != nil {
		return nil, err
	}

	err = h.DHCP.Decline(ctx, requestedIP, pkt.HardwareAddr)
	if err != nil {
		return nil, err
	}

	log.Info("dhcp: marked address as declined", addPacketLog(pkt, map[string]interface{}{
		"requested": requestedIP,
	}))
	return nil, errNoAction
}
