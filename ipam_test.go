package sabakan

import (
	"reflect"
	"testing"
)

func testGenerateIP(t *testing.T) {
	t.Parallel()

	cases := []struct {
		machine       *Machine
		nodeAddresses map[string]string
		bmcAddresses  []string
	}{
		{
			&Machine{
				Serial:      "1234",
				Rack:        1,
				IndexInRack: 3,
			},
			map[string]string{
				"node0": "10.69.0.195",
				"node1": "10.69.1.3",
				"node2": "10.69.1.67",
			},
			[]string{
				"10.72.17.35",
			},
		},
		{
			&Machine{
				Serial:      "5678",
				Rack:        0,
				IndexInRack: 5,
			},
			map[string]string{
				"node0": "10.69.0.5",
				"node1": "10.69.0.69",
				"node2": "10.69.0.133",
			},
			[]string{
				"10.72.17.5",
			},
		},
	}
	config := IPAMConfig{
		MaxNodesInRack:  28,
		NodeIPv4Offset:  "10.69.0.0/26",
		NodeRackShift:   6,
		NodeIndexOffset: 3,
		BMCIPv4Offset:   "10.72.17.0/27",
		BMCRackShift:    5,
		NodeIPPerNode:   3,
		BMCIPPerNode:    1,
	}

	for _, c := range cases {
		config.GenerateIP(c.machine)

		if len(c.machine.Network) != int(config.NodeIPPerNode) {
			t.Fatal("wrong number of networks")
		}
		for k, v := range c.nodeAddresses {
			if c.machine.Network[k].IPv4[0] != v {
				t.Error("wrong IP Address: ", c.machine.Network[k].IPv4[0])
			}
		}
		if !reflect.DeepEqual(c.machine.BMC.IPv4, c.bmcAddresses) {
			t.Errorf("wrong IP Address: %v", c.machine.BMC.IPv4)
		}
	}
}

func TestIPAM(t *testing.T) {
	t.Run("GenerateIP", testGenerateIP)
}
