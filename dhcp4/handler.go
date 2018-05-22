package dhcp4

import (
	"context"
	"net"

	"github.com/cybozu-go/sabakan"
	"go.universe.tf/netboot/dhcp4"
)

type Handler interface {
	ServeDHCP(ctx context.Context, pkt *dhcp4.Packet, intf *net.Interface) (*dhcp4.Packet, error)
}

type DHCPHandler struct {
	Model sabakan.Model
}

func (h DHCPHandler) ServeDHCP(ctx context.Context, pkt *dhcp4.Packet, intf *net.Interface) (*dhcp4.Packet, error) {
	return pkt, nil
}
