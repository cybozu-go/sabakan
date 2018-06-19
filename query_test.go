package sabakan

import "testing"

func TestMatch(t *testing.T) {
	t.Parallel()

	cases := []struct {
		q Query
		m *Machine
		b bool
	}{
		{Query{Serial: "1234"}, NewMachine(MachineSpec{Serial: "1234"}), true},
		{Query{Product: "R630"}, NewMachine(MachineSpec{Product: "R630"}), true},
		{Query{Datacenter: "us"}, NewMachine(MachineSpec{Datacenter: "us"}), true},
		{Query{Rack: "1"}, NewMachine(MachineSpec{Rack: 1}), true},
		{Query{Role: "boot"}, NewMachine(MachineSpec{Role: "boot"}), true},
		{Query{IPv4: "10.20.30.40"}, NewMachine(MachineSpec{IPv4: []string{"10.20.30.40", "10.21.30.40"}}), true},
		{Query{IPv6: "aa::ff"}, NewMachine(MachineSpec{IPv6: []string{"aa::ff", "bb::ff"}}), true},
		{Query{Product: "R630", Datacenter: "us"}, NewMachine(MachineSpec{Product: "R630", Datacenter: "us"}), true},
		{Query{State: "healthy"}, NewMachine(MachineSpec{}), true},

		{Query{Product: "R630", Datacenter: "jp"}, NewMachine(MachineSpec{Product: "R630", Datacenter: "us"}), false},
		{Query{Product: "R730", Datacenter: "us"}, NewMachine(MachineSpec{Product: "R630", Datacenter: "us"}), false},
		{Query{IPv4: "10.20.30.40"}, NewMachine(MachineSpec{IPv4: []string{"10.21.30.40", "10.22.30.40"}}), false},
		{Query{IPv6: "aa::ff"}, NewMachine(MachineSpec{IPv6: []string{"bb::ff", "cc::ff"}}), false},
		{Query{IPv4: "10.20.30.40"}, NewMachine(MachineSpec{}), false},
		{Query{IPv6: "aa::ff"}, NewMachine(MachineSpec{}), false},
		{Query{State: "dead"}, NewMachine(MachineSpec{}), false},
	}

	for _, c := range cases {
		if b := c.q.Match(c.m); b != c.b {
			t.Errorf("wrong match for %#v", c.q)
		}
	}
}
