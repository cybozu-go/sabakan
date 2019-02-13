package sabakan

import (
	"net"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var (
	testIPAMConfig = &IPAMConfig{
		MaxNodesInRack:    28,
		NodeIPv4Pool:      "10.69.0.0/20",
		NodeIPv4Offset:    "",
		NodeRangeSize:     6,
		NodeRangeMask:     26,
		NodeIPPerNode:     3,
		NodeIndexOffset:   3,
		NodeGatewayOffset: 1,
		BMCIPv4Pool:       "10.72.16.0/20",
		BMCIPv4Offset:     "0.0.1.0",
		BMCRangeSize:      5,
		BMCRangeMask:      20,
		BMCGatewayOffset:  1,
	}
)

func testGenerateIP(t *testing.T) {
	t.Parallel()

	cases := []struct {
		machine       *Machine
		nodeAddresses []string
		nic0          NICConfig
		bmc           NICConfig
	}{
		{
			NewMachine(MachineSpec{
				Serial:      "1234",
				Rack:        1,
				IndexInRack: 3,
			}),
			[]string{
				"10.69.0.195",
				"10.69.1.3",
				"10.69.1.67",
			},
			NICConfig{
				"10.69.0.195",
				"255.255.255.192",
				26,
				"10.69.0.193",
			},
			NICConfig{
				"10.72.17.35",
				"255.255.240.0",
				20,
				"10.72.16.1",
			},
		},
		{
			NewMachine(MachineSpec{
				Serial:      "5678",
				Rack:        0,
				IndexInRack: 5,
			}),
			[]string{
				"10.69.0.5",
				"10.69.0.69",
				"10.69.0.133",
			},
			NICConfig{
				"10.69.0.5",
				"255.255.255.192",
				26,
				"10.69.0.1",
			},
			NICConfig{
				"10.72.17.5",
				"255.255.240.0",
				20,
				"10.72.16.1",
			},
		},
	}

	for _, c := range cases {
		testIPAMConfig.GenerateIP(c.machine)
		spec := c.machine.Spec
		info := c.machine.Info

		if len(spec.IPv4) != int(testIPAMConfig.NodeIPPerNode) {
			t.Fatal("wrong number of networks")
		}
		if !reflect.DeepEqual(c.nodeAddresses, spec.IPv4) {
			t.Error("wrong IP addresses: ", spec.IPv4)
		}

		if len(info.Network.IPv4) != int(testIPAMConfig.NodeIPPerNode) {
			t.Fatal("too few NIC config")
		}
		if !cmp.Equal(info.Network.IPv4[0], c.nic0) {
			t.Error("unexpected NIC#0 config", cmp.Diff(info.Network.IPv4[0], c.nic0))
		}
		if !cmp.Equal(info.BMC.IPv4, c.bmc) {
			t.Error("unexpected BMC NIC config", cmp.Diff(info.BMC.IPv4, c.bmc))
		}
	}
}

func testLeaseRange(t *testing.T) {
	t.Parallel()

	r := testIPAMConfig.LeaseRange(net.ParseIP("10.68.10.20"))
	if r != nil {
		t.Error("lease range for 10.68.10.20 should be nil")
	}

	r = testIPAMConfig.LeaseRange(net.ParseIP("10.69.10.20"))
	if r == nil {
		t.Fatal("lease range for 10.69.10.20 must not be nil")
	}

	if r.BeginAddress.String() != "10.69.10.32" {
		t.Error(`r.BeginAddress.String() != "10.69.10.32:"`, r.BeginAddress.String())
	}
	if r.Count != 31 {
		t.Error(`r.Count != 31:`, r.Count)
	}
	if r.IP(3).String() != "10.69.10.35" {
		t.Error(`r.IP(3).String() != "10.69.10.35"`, r.IP(3).String())
	}
	if r.Key() != "10.69.10.32" {
		t.Error(`r.Key() != "10.69.10.32:"`, r.Key())
	}
}

func TestIPAM(t *testing.T) {
	t.Run("GenerateIP", testGenerateIP)
	t.Run("LeaseRange", testLeaseRange)
}
