package dhcpd

import (
	"context"
	"encoding/binary"
	"errors"
	"net"

	"github.com/cybozu-go/log"
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

func leaseSeconds(m sabakan.Model) ([]byte, error) {
	config, err := m.DHCP.GetConfig()
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(config.LeaseMinutes)*60)
	return buf, nil
}
