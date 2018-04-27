package sabakan

import "fmt"

// Query is an URL query
type Query struct {
	Serial     string
	Product    string
	Datacenter string
	Rack       string
	Role       string
	IPv4       string
	IPv6       string
}

func (q *Query) Match(m *Machine) bool {
	if len(q.Serial) > 0 && q.Serial != m.Serial {
		return false
	}
	if len(q.IPv4) > 0 {
		ok := false
		for _, n := range m.Network {
			if n.hasIPv4(q.IPv4) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	if len(q.IPv6) > 0 {
		ok := false
		for _, n := range m.Network {
			if n.hasIPv6(q.IPv6) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	if len(q.Product) > 0 && q.Product != m.Product {
		return false
	}
	if len(q.Datacenter) > 0 && q.Datacenter != m.Datacenter {
		return false
	}
	if len(q.Rack) > 0 && q.Rack != fmt.Sprint(m.Rack) {
		return false
	}
	if len(q.Role) > 0 && q.Role != m.Role {
		return false
	}

	return true
}
