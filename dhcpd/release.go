package dhcpd

import (
	"context"

	"github.com/cybozu-go/log"
	"go.universe.tf/netboot/dhcp4"
)

func (h DHCPHandler) handleRelease(ctx context.Context, pkt *dhcp4.Packet, intf Interface) (*dhcp4.Packet, error) {
	serverAddr, err := getIPv4AddrForInterface(intf)
	if err != nil {
		return nil, err
	}

	serverIdentifier, err := pkt.Options.IP(dhcp4.OptServerIdentifier)
	if err != nil {
		return nil, err
	}

	if !serverAddr.Equal(serverIdentifier) {
		log.Warn("dhcp: requested release with inconsistent server id", addPacketLog(pkt, map[string]interface{}{
			optionLogKey(dhcp4.OptServerIdentifier): serverIdentifier,
		}))
		return nil, errNotChosen
	}

	err = h.DHCP.Release(ctx, pkt.ClientAddr, pkt.HardwareAddr)
	if err != nil {
		return nil, err
	}

	log.Info("dhcp: released address", addPacketLog(pkt, map[string]interface{}{
		pktCiaddr: pkt.ClientAddr,
	}))
	return nil, errNoAction
}
