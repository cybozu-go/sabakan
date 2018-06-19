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
	BMCType    string
}

// Match returns true if all non-empty fields matches Machine
func (q *Query) Match(m *MachineSpec) bool {
	if len(q.Serial) > 0 && q.Serial != m.Serial {
		return false
	}
	if len(q.IPv4) > 0 {
		ok := false
		for _, ip := range m.IPv4 {
			if ip == q.IPv4 {
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
		for _, ip := range m.IPv6 {
			if ip == q.IPv6 {
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
	if len(q.BMCType) > 0 && q.BMCType != m.BMC.Type {
		return false
	}

	return true
}

// IsEmpty returns true if query is empty
func (q *Query) IsEmpty() bool {
	return q.Serial == "" && q.Product == "" && q.Datacenter == "" && q.Rack == "" &&
		q.Role == "" && q.IPv4 == "" && q.BMCType == ""
}

// QueryBySerial create Query by serial
func QueryBySerial(serial string) *Query {
	return &Query{
		Serial: serial,
	}
}

// QueryByIPv4 create Query by IPv4 address
func QueryByIPv4(ipv4 string) *Query {
	return &Query{
		IPv4: ipv4,
	}
}
