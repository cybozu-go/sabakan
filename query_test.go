package sabakan

import "testing"

func TestMatch(t *testing.T) {
	t.Parallel()

	cases := []struct {
		query   Query
		machine *Machine
		matched bool
	}{
		{Query{"serial": "1234"}, NewMachine(MachineSpec{Serial: "1234"}), true},
		{Query{"labels": "product=R630"}, NewMachine(MachineSpec{Labels: map[string]string{"product": "R630"}}), true},
		{Query{"labels": "datacenter=us"}, NewMachine(MachineSpec{Labels: map[string]string{"datacenter": "us"}}), true},
		{Query{"rack": "1"}, NewMachine(MachineSpec{Rack: 1}), true},
		{Query{"rack": "1,2"}, NewMachine(MachineSpec{Rack: 1}), true},
		{Query{"role": "boot"}, NewMachine(MachineSpec{Role: "boot"}), true},
		{Query{"role": "boot,worker"}, NewMachine(MachineSpec{Role: "boot"}), true},
		{Query{"ipv4": "10.20.30.40"}, NewMachine(MachineSpec{IPv4: []string{"10.20.30.40", "10.21.30.40"}}), true},
		{Query{"ipv4": "10.20.30.40,10.21.30.40"}, NewMachine(MachineSpec{IPv4: []string{"10.20.30.40", "10.21.30.40"}}), true},
		{Query{"ipv6": "aa::ff"}, NewMachine(MachineSpec{IPv6: []string{"aa::ff", "bb::ff"}}), true},
		{Query{"ipv6": "aa::ff,bb::ff"}, NewMachine(MachineSpec{IPv6: []string{"aa::ff", "bb::ff"}}), true},
		{Query{"labels": "product=R630,datacenter=us"}, NewMachine(MachineSpec{Labels: map[string]string{"product": "R630", "datacenter": "us"}}), true},
		{Query{"state": "uninitialized"}, NewMachine(MachineSpec{}), true},
		{Query{"bmc-type": "iDRAC-9"}, NewMachine(MachineSpec{BMC: MachineBMC{Type: "iDRAC-9"}}), true},
		{Query{"bmc-type": "iDRAC-9,IPMI-1.0"}, NewMachine(MachineSpec{BMC: MachineBMC{Type: "iDRAC-9"}}), true},
		{Query{"labels": "product=R630"}, NewMachine(MachineSpec{Labels: map[string]string{"product": "R630", "datacenter": "jp"}}), true},
		{Query{"labels": "product=R630,datacenter=jp"}, NewMachine(MachineSpec{Labels: map[string]string{"product": "R630"}}), false},
		{Query{"labels": "product=R630,datacenter=jp"}, NewMachine(MachineSpec{Labels: map[string]string{"product": "R630", "datacenter": "us"}}), false},
		{Query{"labels": "product=R730,datacenter=us"}, NewMachine(MachineSpec{Labels: map[string]string{"product": "R630", "datacenter": "us"}}), false},
		{Query{"ipv4": "10.20.30.40"}, NewMachine(MachineSpec{IPv4: []string{"10.21.30.40", "10.22.30.40"}}), false},
		{Query{"ipv6": "aa::ff"}, NewMachine(MachineSpec{IPv6: []string{"bb::ff", "cc::ff"}}), false},
		{Query{"ipv4": "10.20.30.40"}, NewMachine(MachineSpec{}), false},
		{Query{"ipv6": "aa::ff"}, NewMachine(MachineSpec{}), false},
		{Query{"state": "unreachable"}, NewMachine(MachineSpec{}), false},
		{Query{"bmc-type": "IPMI-1.0"}, NewMachine(MachineSpec{BMC: MachineBMC{Type: "iDRAC-9"}}), false},
		{Query{"without-serial": "1234"}, NewMachine(MachineSpec{Serial: "1234"}), false},
		{Query{"without-serial": "1234,5678"}, NewMachine(MachineSpec{Serial: "1234"}), false},
		{Query{"without-labels": "product=R630"}, NewMachine(MachineSpec{Labels: map[string]string{"product": "R630"}}), false},
		{Query{"without-labels": "datacenter=us"}, NewMachine(MachineSpec{Labels: map[string]string{"datacenter": "us"}}), false},
		{Query{"without-rack": "1"}, NewMachine(MachineSpec{Rack: 1}), false},
		{Query{"without-rack": "1,2"}, NewMachine(MachineSpec{Rack: 1}), false},
		{Query{"without-role": "boot"}, NewMachine(MachineSpec{Role: "boot"}), false},
		{Query{"without-role": "boot,worker"}, NewMachine(MachineSpec{Role: "boot"}), false},
		{Query{"without-ipv4": "10.20.30.40"}, NewMachine(MachineSpec{IPv4: []string{"10.20.30.40", "10.21.30.40"}}), false},
		{Query{"without-ipv4": "10.20.30.40,10.21.30.40"}, NewMachine(MachineSpec{IPv4: []string{"10.20.30.40", "10.21.30.40"}}), false},
		{Query{"without-ipv6": "aa::ff"}, NewMachine(MachineSpec{IPv6: []string{"aa::ff", "bb::ff"}}), false},
		{Query{"without-ipv6": "aa::ff,bb::ff"}, NewMachine(MachineSpec{IPv6: []string{"aa::ff", "bb::ff"}}), false},
		{Query{"without-labels": "product=R630,datacenter=us"}, NewMachine(MachineSpec{Labels: map[string]string{"product": "R630", "datacenter": "us"}}), false},
		{Query{"without-state": "uninitialized"}, NewMachine(MachineSpec{}), false},
		{Query{"without-state": "uninitialized,retired"}, NewMachine(MachineSpec{}), false},
		{Query{"without-bmc-type": "iDRAC-9"}, NewMachine(MachineSpec{BMC: MachineBMC{Type: "iDRAC-9"}}), false},
		{Query{"without-bmc-type": "iDRAC-9,IPMI-1.0"}, NewMachine(MachineSpec{BMC: MachineBMC{Type: "iDRAC-9"}}), false},
		{Query{"without-labels": "product=R630"}, NewMachine(MachineSpec{Labels: map[string]string{"product": "R630", "datacenter": "jp"}}), false},
		{Query{"without-labels": "product=R630,datacenter=jp"}, NewMachine(MachineSpec{Labels: map[string]string{"product": "R630"}}), true},
		{Query{"without-labels": "product=R630,datacenter=jp"}, NewMachine(MachineSpec{Labels: map[string]string{"product": "R630", "datacenter": "us"}}), true},
		{Query{"without-labels": "product=R730,datacenter=us"}, NewMachine(MachineSpec{Labels: map[string]string{"product": "R630", "datacenter": "us"}}), true},
		{Query{"without-ipv4": "10.20.30.40"}, NewMachine(MachineSpec{IPv4: []string{"10.21.30.40", "10.22.30.40"}}), true},
		{Query{"without-ipv6": "aa::ff"}, NewMachine(MachineSpec{IPv6: []string{"bb::ff", "cc::ff"}}), true},
		{Query{"without-ipv4": "10.20.30.40"}, NewMachine(MachineSpec{}), true},
		{Query{"without-ipv6": "aa::ff"}, NewMachine(MachineSpec{}), true},
		{Query{"without-state": "unreachable"}, NewMachine(MachineSpec{}), true},
		{Query{"without-bmc-type": "IPMI-1.0"}, NewMachine(MachineSpec{BMC: MachineBMC{Type: "iDRAC-9"}}), true},
	}

	for _, c := range cases {
		matched, err := c.query.Match(c.machine)
		if err != nil {
			t.Errorf("error returned for %#v: %v", c.query, err)
		} else if matched != c.matched {
			t.Errorf("wrong match for %#v", c.query)
		}
	}
}

func TestIsEmpty(t *testing.T) {
	blanks := []Query{{}, {"serial": "", "role": ""}}
	for _, q := range blanks {
		if !q.IsEmpty() {
			t.Errorf("q.IsEmpty()")
		}
	}

	presents := []Query{{"user-filed": "hello"}, {"serial": "1234", "role": ""}}
	for _, q := range presents {
		if q.IsEmpty() {
			t.Errorf("!q.IsEmpty()")
		}
	}
}
