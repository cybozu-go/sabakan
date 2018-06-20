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
	State      string
}

// Match returns true if all non-empty fields matches Machine
func (q *Query) Match(m *Machine) bool {
	if len(q.Serial) > 0 && q.Serial != m.Spec.Serial {
		return false
	}
	if len(q.IPv4) > 0 {
		ok := false
		for _, ip := range m.Spec.IPv4 {
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
		for _, ip := range m.Spec.IPv6 {
			if ip == q.IPv6 {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	if len(q.Product) > 0 && q.Product != m.Spec.Product {
		return false
	}
	if len(q.Datacenter) > 0 && q.Datacenter != m.Spec.Datacenter {
		return false
	}
	if len(q.Rack) > 0 && q.Rack != fmt.Sprint(m.Spec.Rack) {
		return false
	}
	if len(q.Role) > 0 && q.Role != m.Spec.Role {
		return false
	}
	if len(q.BMCType) > 0 && q.BMCType != m.Spec.BMC.Type {
		return false
	}
	if len(q.State) > 0 && q.State != m.Status.State.String() {
		return false
	}

	return true
}

// IsEmpty returns true if query is empty
func (q *Query) IsEmpty() bool {
	return q.Serial == "" && q.Product == "" && q.Datacenter == "" && q.Rack == "" &&
		q.Role == "" && q.IPv4 == "" && q.BMCType == "" && q.State == ""
}
