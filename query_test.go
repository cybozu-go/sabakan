package sabakan

import "testing"

func TestMatch(t *testing.T) {
	t.Parallel()

	cases := []struct {
		q Query
		m Machine
		b bool
	}{
		{Query{Serial: "1234"}, Machine{Serial: "1234"}, true},
		{Query{Product: "R630"}, Machine{Product: "R630"}, true},
		{Query{Datacenter: "us"}, Machine{Datacenter: "us"}, true},
		{Query{Rack: "1"}, Machine{Rack: 1}, true},
		{Query{Role: "boot"}, Machine{Role: "boot"}, true},
		{Query{IPv4: "10.20.30.40"}, Machine{IPv4: []string{"10.20.30.40", "10.21.30.40"}}, true},
		{Query{IPv6: "aa::ff"}, Machine{IPv6: []string{"aa::ff", "bb::ff"}}, true},
		{Query{Product: "R630", Datacenter: "us"}, Machine{Product: "R630", Datacenter: "us"}, true},
		{Query{Product: "R630", Datacenter: "jp"}, Machine{Product: "R630", Datacenter: "us"}, false},
		{Query{Product: "R730", Datacenter: "us"}, Machine{Product: "R630", Datacenter: "us"}, false},

		{Query{IPv4: "10.20.30.40"}, Machine{IPv4: []string{"10.21.30.40", "10.22.30.40"}}, false},
		{Query{IPv6: "aa::ff"}, Machine{IPv6: []string{"bb::ff", "cc::ff"}}, false},
		{Query{IPv4: "10.20.30.40"}, Machine{}, false},
		{Query{IPv6: "aa::ff"}, Machine{}, false},
	}

	for _, c := range cases {
		if b := c.q.Match(&c.m); b != c.b {
			t.Errorf("wrong match for %#v", c.q)
		}
	}
}
