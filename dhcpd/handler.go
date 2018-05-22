package dhcpd

import (
	"context"
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
		//	case dhcp4.MsgRequest:
		//		return h.handleRequest(ctx, pkt, intf)
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
