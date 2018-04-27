package sabakan

import (
	"errors"
	"net"
)

// IPAMConfig is structure of the sabakan option
type IPAMConfig struct {
	NodeIPv4Offset string `json:"node-ipv4-offset"`
	NodeRackShift  uint   `json:"node-rack-shift"`
	BMCIPv4Offset  string `json:"bmc-ipv4-offset"`
	BMCRackShift   uint   `json:"bmc-rack-shift"`
	NodeIPPerNode  uint   `json:"node-ip-per-node"`
	BMCIPPerNode   uint   `json:"bmc-ip-per-node"`
}

// Validate validates configurations
func (c *IPAMConfig) Validate() error {
	if _, _, err := net.ParseCIDR(c.NodeIPv4Offset); err != nil {
		return errors.New("invalid node-ipv4-offset")
	}
	if c.NodeRackShift == 0 {
		return errors.New("node-rack-shift must not be zero")
	}
	if _, _, err := net.ParseCIDR(c.BMCIPv4Offset); err != nil {
		return errors.New("invalid bmc-ipv4-offset")
	}
	if c.BMCRackShift == 0 {
		return errors.New("bmc-rack-shift must not be zero")
	}
	if c.NodeIPPerNode == 0 {
		return errors.New("node-ip-per-node must not be zero")
	}
	if c.BMCIPPerNode == 0 {
		return errors.New("bmc-ip-per-node must not be zero")
	}
	return nil
}
