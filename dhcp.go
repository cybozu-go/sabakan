package sabakan

import (
	"errors"
	"net"

	"github.com/cybozu-go/netutil"
)

// DHCPConfig is a set of DHCP configurations.
type DHCPConfig struct {
	GatewayOffset uint `json:"gateway-offset"`
}

// GatewayAddress returns a gateway address for the given node address
func (c *DHCPConfig) GatewayAddress(addr *net.IPNet) *net.IPNet {
	a := netutil.IP4ToInt(addr.IP.Mask(addr.Mask))
	a += uint32(c.GatewayOffset)
	return &net.IPNet{
		IP:   netutil.IntToIP4(a),
		Mask: addr.Mask,
	}
}

// Validate validates configurations
func (c *DHCPConfig) Validate() error {
	if c.GatewayOffset == 0 {
		return errors.New("gateway-offset must not be zero")
	}

	return nil
}
