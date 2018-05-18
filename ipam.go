package sabakan

import (
	"errors"
	"fmt"
	"net"

	"github.com/cybozu-go/netutil"
)

// IPAMConfig is a set of IPAM configurations.
type IPAMConfig struct {
	MaxNodesInRack  uint   `json:"max-nodes-in-rack"`
	NodeIPv4Pool    string `json:"node-ipv4-pool"`
	NodeRangeSize   uint   `json:"node-ipv4-range-size"`
	NodeRangeMask   uint   `json:"node-ipv4-range-mask"`
	NodeIPPerNode   uint   `json:"node-ip-per-node"`
	NodeIndexOffset uint   `json:"node-index-offset"`
	BMCIPv4Pool     string `json:"bmc-ipv4-pool"`
	BMCRangeSize    uint   `json:"bmc-ipv4-range-size"`
	BMCRangeMask    uint   `json:"bmc-ipv4-range-mask"`
}

// Validate validates configurations
func (c *IPAMConfig) Validate() error {
	if c.MaxNodesInRack == 0 {
		return errors.New("max-nodes-in-rack must not be zero")
	}

	ip, ipNet, err := net.ParseCIDR(c.NodeIPv4Pool)
	if err != nil {
		return errors.New("invalid node-ipv4-pool")
	}
	if !ip.Equal(ipNet.IP) {
		return errors.New("host part of node-ipv4-pool must be cleared")
	}
	if c.NodeRangeSize == 0 {
		return errors.New("node-ipv4-range-size must not be zero")
	}
	if c.NodeRangeMask < 8 || 32 < c.NodeRangeMask {
		return errors.New("invalid node-ipv4-range-mask")
	}
	if c.NodeIPPerNode == 0 {
		return errors.New("node-ip-per-node must not be zero")
	}
	if c.NodeIndexOffset == 0 {
		return errors.New("node-index-offset must not be zero")
	}

	ip, ipNet, err = net.ParseCIDR(c.BMCIPv4Pool)
	if err != nil {
		return errors.New("invalid bmc-ipv4-pool")
	}
	if !ip.Equal(ipNet.IP) {
		return errors.New("host part of bmc-ipv4-pool must be cleared")
	}
	if c.BMCRangeSize == 0 {
		return errors.New("bmc-ipv4-range-size must not be zero")
	}
	if c.BMCRangeMask < 8 || 32 < c.BMCRangeMask {
		return errors.New("invalid bmc-ipv4-range-mask")
	}

	return nil
}

// GenerateIP generates IP addresses for a machine.
// Generated IP addresses are stored in mc.
func (c *IPAMConfig) GenerateIP(mc *Machine) {
	// IP addresses are calculated as following (LRN=Logical Rack Number):
	// node0: INET_NTOA(INET_ATON(NodeIPv4Pool) + (2^NodeRangeSize * NodeIPPerNode * LRN) + index-in-rack)
	// node1: INET_NTOA(INET_ATON(NodeIPv4Pool) + (2^NodeRangeSize * NodeIPPerNode * LRN) + index-in-rack + 2^NodeRangeSize)
	// node2: INET_NTOA(INET_ATON(NodeIPv4Pool) + (2^NodeRangeSize * NodeIPPerNode * LRN) + index-in-rack + 2^NodeRangeSize * 2)
	// BMC: INET_NTOA(INET_ATON(BMCIPv4Pool) + (2^BMCRangeSize * LRN) + index-in-rack)

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

	ips := calc(c.NodeIPv4Pool, c.NodeRangeSize, c.NodeIPPerNode, mc.Rack, mc.IndexInRack)
	res := map[string]MachineNetwork{}
	for i := 0; i < int(c.NodeIPPerNode); i++ {
		name := fmt.Sprintf("node%d", i)
		res[name] = MachineNetwork{
			IPv4: []string{ips[i].String()},
		}
	}
	mc.Network = res

	bmcIPs := calc(c.BMCIPv4Pool, c.BMCRangeSize, 1, mc.Rack, mc.IndexInRack)
	mc.BMC.IPv4 = bmcIPs[0].String()
}

// LeaseRange is a range of IP addresses for DHCP lease.
type LeaseRange struct {
	BeginAddress net.IP
	Count        int
}

// LeaseRange returns a LeaseRange for the interface that receives DHCP requests.
// If no range can be assigned, this returns nil.
func (c *IPAMConfig) LeaseRange(ifaddr net.IP) *LeaseRange {
	ip1, _, _ := net.ParseCIDR(c.NodeIPv4Pool)
	nip1 := netutil.IP4ToInt(ip1)
	nip2 := netutil.IP4ToInt(ifaddr)
	if nip2 <= nip1 {
		return nil
	}

	// Given these configurations,
	//   MaxNodesInRack  = 28
	//   NodeRangeSize   = 6
	//   NodeIndexOffset = 3
	//
	// The lease range will start at offset 32, and ends at 62 (64 - 1 - 1).
	// Therefore the available lease IP address count is 31.

	rangeSize := uint32(1 << c.NodeRangeSize)
	offset := uint32(c.NodeIndexOffset + c.MaxNodesInRack + 1)

	ranges := (nip2 - nip1) / rangeSize
	rangeStart := nip1 + rangeSize*ranges + uint32(c.NodeIndexOffset+c.MaxNodesInRack+1)
	startIP := netutil.IntToIP4(rangeStart)
	count := (rangeSize - 2) - offset + 1
	return &LeaseRange{
		BeginAddress: startIP,
		Count:        int(count),
	}
}
