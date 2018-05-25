package dhcpd

import (
	"context"

	"github.com/cybozu-go/log"
	"go.universe.tf/netboot/dhcp4"
)

func (h DHCPHandler) handleDecline(ctx context.Context, pkt *dhcp4.Packet, intf Interface) (*dhcp4.Packet, error) {
	log.Info("received", map[string]interface{}{
		"intf":      intf.Name(),
		"type":      "DHCPDECLINE",
		"xid":       pkt.TransactionID,
		"broadcast": pkt.Broadcast,
		"chaddr":    pkt.HardwareAddr,
	})
	log.Debug("received", getOptionsLog(pkt))

	serverAddr, err := getIPv4AddrForInterface(intf)
	if err != nil {
		return nil, err
	}

	serverIdentifier, err := pkt.Options.IP(dhcp4.OptServerIdentifier)
	if err != nil {
		return nil, err
	}

	if !serverAddr.Equal(serverIdentifier) {
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
	return nil, errNoAction
}
