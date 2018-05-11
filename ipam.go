package sabakan

import (
	"errors"
	"fmt"
	"net"

	"github.com/cybozu-go/netutil"
)

// IPAMConfig is structure of the sabakan option
type IPAMConfig struct {
	MaxRacks       uint   `json:"max-racks"`
	MaxNodesInRack uint   `json:"max-nodes-in-rack"`
	NodeIPv4Offset string `json:"node-ipv4-offset"`
	NodeRackShift  uint   `json:"node-rack-shift"`
	NodeIPPerNode  uint   `json:"node-ip-per-node"`
	BMCIPv4Offset  string `json:"bmc-ipv4-offset"`
	BMCRackShift   uint   `json:"bmc-rack-shift"`
	BMCIPPerNode   uint   `json:"bmc-ip-per-node"`
}

// Validate validates configurations
func (c *IPAMConfig) Validate() error {
	if c.MaxRacks == 0 {
		return errors.New("max-racks must not be zero")
	}
	if c.MaxNodesInRack == 0 {
		return errors.New("max-nodes-in-rack must not be zero")
	}
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

// GenerateIP generates IP addresses for a machine.
// Generated IP addresses are stored in mc.
func (c *IPAMConfig) GenerateIP(mc *Machine) {
	// IP addresses are calculated as following:
	// node0: INET_NTOA(INET_ATON(NodeIPv4Offset) + (2^NodeRackShift * NodeIPPerNode * rack-number) + node-index-in-rack)
	// node1: INET_NTOA(INET_ATON(NodeIPv4Offset) + (2^NodeRackShift * NodeIPPerNode * rack-number) + node-index-in-rack + 2^NodeRackShift)
	// node2: INET_NTOA(INET_ATON(NodeIPv4Offset) + (2^NodeRackShift * NodeIPPerNode * rack-number) + node-index-in-rack + 2^NodeRackShift * 2)
	// BMC: INET_NTOA(INET_ATON(BMCIPv4Offset) + (2^BMCRackShift * BMCIPPerNode * rack-number) + node-index-in-rack)

	calc := func(cidr string, shift uint, numip uint, lrn uint, nodeIndex uint) []net.IP {
		result := make([]net.IP, numip)

		offset, _, _ := net.ParseCIDR(cidr)
		a := netutil.IP4ToInt(offset)
		su := uint(1) << shift
		for i := uint(0); i < numip; i++ {
			ip := netutil.IntToIP4(a + uint32(su*numip*lrn+nodeIndex+i*su))
			result[i] = ip
		}
		return result
	}

	ips := calc(c.NodeIPv4Offset, c.NodeRackShift, c.NodeIPPerNode, mc.Rack, mc.NodeIndexInRack)
	res := map[string]MachineNetwork{}
	for i := 0; i < int(c.NodeIPPerNode); i++ {
		name := fmt.Sprintf("node%d", i)
		res[name] = MachineNetwork{
			IPv4: []string{ips[i].String()},
		}
	}
	mc.Network = res

	bmcIPs := calc(c.BMCIPv4Offset, c.BMCRackShift, c.BMCIPPerNode, mc.Rack, mc.NodeIndexInRack)
	for _, ip := range bmcIPs {
		mc.BMC.IPv4 = append(mc.BMC.IPv4, ip.String())
	}
}
