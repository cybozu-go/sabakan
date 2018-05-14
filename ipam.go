package sabakan

import (
	"errors"
	"fmt"
	"net"

	"github.com/cybozu-go/netutil"
)

// IPAMConfig is structure of the sabakan option
type IPAMConfig struct {
	MaxNodesInRack  uint   `json:"max-nodes-in-rack"`
	NodeIPv4Offset  string `json:"node-ipv4-offset"`
	NodeRackShift   uint   `json:"node-rack-shift"`
	NodeIndexOffset uint   `json:"node-index-offset"`
	NodeIPPerNode   uint   `json:"node-ip-per-node"`
	BMCIPv4Offset   string `json:"bmc-ipv4-offset"`
	BMCRackShift    uint   `json:"bmc-rack-shift"`
	BMCIPPerNode    uint   `json:"bmc-ip-per-node"`
}

// Validate validates configurations
func (c *IPAMConfig) Validate() error {
	if c.MaxNodesInRack == 0 {
		return errors.New("max-nodes-in-rack must not be zero")
	}

	ip, ipNet, err := net.ParseCIDR(c.NodeIPv4Offset)
	if err != nil {
		return errors.New("invalid node-ipv4-offset")
	}
	if !ip.Equal(ipNet.IP) {
		return errors.New("host part of node-ipv4-offset must be 0s")
	}
	if c.NodeRackShift == 0 {
		return errors.New("node-rack-shift must not be zero")
	}

	ip, ipNet, err = net.ParseCIDR(c.BMCIPv4Offset)
	if err != nil {
		return errors.New("invalid bmc-ipv4-offset")
	}
	if !ip.Equal(ipNet.IP) {
		return errors.New("host part of bmc-ipv4-offset must be 0s")
	}
	if c.BMCRackShift == 0 {
		return errors.New("bmc-rack-shift must not be zero")
	}

	if c.NodeIndexOffset == 0 {
		return errors.New("node-index-offset must not be zero")
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
	// IP addresses are calculated as following (LRN=Logical Rack Number):
	// node0: INET_NTOA(INET_ATON(NodeIPv4Offset) + (2^NodeRackShift * NodeIPPerNode * LRN) + index-in-rack)
	// node1: INET_NTOA(INET_ATON(NodeIPv4Offset) + (2^NodeRackShift * NodeIPPerNode * LRN) + index-in-rack + 2^NodeRackShift)
	// node2: INET_NTOA(INET_ATON(NodeIPv4Offset) + (2^NodeRackShift * NodeIPPerNode * LRN) + index-in-rack + 2^NodeRackShift * 2)
	// BMC: INET_NTOA(INET_ATON(BMCIPv4Offset) + (2^BMCRackShift * BMCIPPerNode * LRN) + index-in-rack)

	calc := func(cidr string, shift, numip, lrn, idx uint) []net.IP {
		result := make([]net.IP, numip)

		offset, _, _ := net.ParseCIDR(cidr)
		a := netutil.IP4ToInt(offset)
		su := uint(1) << shift
		for i := uint(0); i < numip; i++ {
			ip := netutil.IntToIP4(a + uint32(su*numip*lrn+idx+i*su))
			result[i] = ip
		}
		return result
	}

	ips := calc(c.NodeIPv4Offset, c.NodeRackShift, c.NodeIPPerNode, mc.Rack, mc.IndexInRack)
	res := map[string]MachineNetwork{}
	for i := 0; i < int(c.NodeIPPerNode); i++ {
		name := fmt.Sprintf("node%d", i)
		res[name] = MachineNetwork{
			IPv4: []string{ips[i].String()},
		}
	}
	mc.Network = res

	bmcIPs := calc(c.BMCIPv4Offset, c.BMCRackShift, c.BMCIPPerNode, mc.Rack, mc.IndexInRack)
	for _, ip := range bmcIPs {
		mc.BMC.IPv4 = append(mc.BMC.IPv4, ip.String())
	}
}
