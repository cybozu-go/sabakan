package dhcpd

import (
	"context"
	"errors"
	"net"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan"
	"go.universe.tf/netboot/dhcp4"
)

// Handler defines interface for dhcp service
type Handler interface {
	ServeDHCP(ctx context.Context, pkt *dhcp4.Packet, intf *net.Interface) (*dhcp4.Packet, error)
}

// DHCPHandler implements Handler
type DHCPHandler struct {
	Model sabakan.Model
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
	return nil, errors.New("unknown message type")
}
