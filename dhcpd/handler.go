package dhcpd

import (
	"context"
	"encoding/binary"
	"errors"
	"net"
	"net/url"
	"path"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/netutil"
	"github.com/cybozu-go/sabakan"
	"go.universe.tf/netboot/dhcp4"
)

// Handler defines an interface for Server.
type Handler interface {
	ServeDHCP(ctx context.Context, pkt *dhcp4.Packet, intf Interface) (*dhcp4.Packet, error)
}

// DHCPHandler is an implementation of Handler using sabakan.Model.
type DHCPHandler struct {
	sabakan.Model
	MyURL *url.URL
}

// ServeDHCP implements Handler interface
func (h DHCPHandler) ServeDHCP(ctx context.Context, pkt *dhcp4.Packet, intf Interface) (*dhcp4.Packet, error) {
	switch pkt.Type {
	case dhcp4.MsgDiscover:
		return h.handleDiscover(ctx, pkt, intf)
	case dhcp4.MsgRequest:
		return h.handleRequest(ctx, pkt, intf)
	case dhcp4.MsgDecline:
		return h.handleDecline(ctx, pkt, intf)
	case dhcp4.MsgRelease:
		return h.handleRelease(ctx, pkt, intf)
	case dhcp4.MsgInform:
		return h.handleInform(ctx, pkt, intf)
	default:
		log.Error("unexpected message type", map[string]interface{}{
			"type": pkt.Type.String(),
		})
	}
	return nil, errUnknownMsgType
}

func getIPv4AddrForInterface(intf Interface) (net.IP, error) {
	addrs, err := intf.Addrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		ipaddr, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		ipaddr4 := ipaddr.IP.To4()
		if ipaddr4 != nil {
			return ipaddr4, nil
		}
	}
	return nil, errors.New("No IPv4 address for " + intf.Name())
}

func flattenIPv4Addresses(addrs []string) []byte {
	buf := make([]byte, len(addrs)*4)
	for i, addr := range addrs {
		copy(buf[i*4:(i+1)*4], net.ParseIP(addr).To4())
	}
	return buf
}

// makeOptions returns dhcp4.Options that includes these common options:
//
// * Subnet Mask (1)
// * Router (3)
// * Domain Name Server (6) (if specified in DHCP config)
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

	// domain name server
	if len(config.DNSServers) > 0 {
		opts[dhcp4.OptDNSServers] = flattenIPv4Addresses(config.DNSServers)
	}

	// lease seconds
	secs := uint32(config.LeaseDuration().Seconds())
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, secs)
	opts[dhcp4.OptLeaseTime] = buf

	return opts, nil
}

func (h DHCPHandler) makeBootAPIURL(p string) string {
	u := *h.MyURL
	u.Path = path.Join("/api/v1/boot", p)
	return u.String()
}
