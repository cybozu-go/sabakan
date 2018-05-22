package dhcpd

import (
	"context"
	"encoding/binary"
	"errors"
	"net"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/netutil"
	"github.com/cybozu-go/sabakan"
	"go.universe.tf/netboot/dhcp4"
)

// Handler defines an interface for Server.
type Handler interface {
	ServeDHCP(ctx context.Context, pkt *dhcp4.Packet, intf *net.Interface) (*dhcp4.Packet, error)
}

// DHCPHandler is an implementation of Handler using sabakan.Model.
type DHCPHandler struct {
	sabakan.Model
}

// ServeDHCP implements Handler interface
func (h DHCPHandler) ServeDHCP(ctx context.Context, pkt *dhcp4.Packet, intf *net.Interface) (*dhcp4.Packet, error) {
	switch pkt.Type {
	case dhcp4.MsgDiscover:
		return h.handleDiscover(ctx, pkt, intf)
	case dhcp4.MsgRequest:
		return h.handleRequest(ctx, pkt, intf)
		//	case dhcp4.MsgDecline:
		//		return h.handleDecline(ctx, pkt, intf)
		//	case dhcp4.MsgRelease:
		//		return h.handleRelease(ctx, pkt, intf)
		//	case dhcp4.MsgInform:
		//		return h.handleInform(ctx, pkt, intf)
	default:
		log.Error("unexpected message type", map[string]interface{}{
			"type": pkt.Type.String(),
		})
	}
	return nil, errUnknownMsgType
}

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

// makeOptions returns dhcp4.Options that includes these common options:
//
// * Subnet Mask (1)
// * Router (3)
// * Lease seconds (51)
func (h DHCPHandler) makeOptions(ciaddr net.IP) (dhcp4.Options, error) {
	ipam, err := h.IPAM.GetConfig()
	if err != nil {
		return nil, err
	}
	config, err := h.DHCP.GetConfig()
	if err != nil {
		return nil, err
	}

	opts := make(dhcp4.Options)

	// subnet mask
	mask := net.CIDRMask(int(ipam.NodeRangeMask), 32)
	opts[dhcp4.OptSubnetMask] = mask

	// default gateway address (router)
	nnet := netutil.IP4ToInt(ciaddr.Mask(mask))
	gw := netutil.IntToIP4(nnet + uint32(config.GatewayOffset)).To4()
	opts[dhcp4.OptRouters] = gw

	// lease seconds
	secs := uint32(config.LeaseDuration().Seconds())
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, secs)
	opts[dhcp4.OptLeaseTime] = buf

	return opts, nil
}
