package sabakan

import "testing"

func TestMatch(t *testing.T) {
	t.Parallel()

	cases := []struct {
		q Query
		m *Machine
		b bool
	}{
		{Query{"serial": "1234"}, NewMachine(MachineSpec{Serial: "1234"}), true},
		{Query{"labels": "product=R630"}, NewMachine(MachineSpec{Labels: map[string]string{"product": "R630"}}), true},
		{Query{"labels": "datacenter=us"}, NewMachine(MachineSpec{Labels: map[string]string{"datacenter": "us"}}), true},
		{Query{"rack": "1"}, NewMachine(MachineSpec{Rack: 1}), true},
		{Query{"role": "boot"}, NewMachine(MachineSpec{Role: "boot"}), true},
		{Query{"ipv4": "10.20.30.40"}, NewMachine(MachineSpec{IPv4: []string{"10.20.30.40", "10.21.30.40"}}), true},
		{Query{"ipv6": "aa::ff"}, NewMachine(MachineSpec{IPv6: []string{"aa::ff", "bb::ff"}}), true},
		{Query{"labels": "product=R630,datacenter=us"}, NewMachine(MachineSpec{Labels: map[string]string{"product": "R630", "datacenter": "us"}}), true},
		{Query{"state": "healthy"}, NewMachine(MachineSpec{}), true},
		{Query{"labels": "product=R630"}, NewMachine(MachineSpec{Labels: map[string]string{"product": "R630", "datacenter": "jp"}}), true},
		{Query{"labels": "product=R630,datacenter=jp"}, NewMachine(MachineSpec{Labels: map[string]string{"product": "R630"}}), false},
		{Query{"labels": "product=R630,datacenter=jp"}, NewMachine(MachineSpec{Labels: map[string]string{"product": "R630", "datacenter": "us"}}), false},
		{Query{"labels": "product=R730,datacenter=us"}, NewMachine(MachineSpec{Labels: map[string]string{"product": "R630", "datacenter": "us"}}), false},
		{Query{"ipv4": "10.20.30.40"}, NewMachine(MachineSpec{IPv4: []string{"10.21.30.40", "10.22.30.40"}}), false},
		{Query{"ipv6": "aa::ff"}, NewMachine(MachineSpec{IPv6: []string{"bb::ff", "cc::ff"}}), false},
		{Query{"ipv4": "10.20.30.40"}, NewMachine(MachineSpec{}), false},
		{Query{"ipv6": "aa::ff"}, NewMachine(MachineSpec{}), false},
		{Query{"state": "dead"}, NewMachine(MachineSpec{}), false},
	}

	for _, c := range cases {
		if b := c.q.Match(c.m); b != c.b {
			t.Errorf("wrong match for %#v", c.q)
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
