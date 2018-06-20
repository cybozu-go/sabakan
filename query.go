package sabakan

import "fmt"

// Query is an URL query
type Query map[string]string

// Match returns true if all non-empty fields matches Machine
func (q Query) Match(m *Machine) bool {
	if serial := q["serial"]; len(serial) > 0 && serial != m.Spec.Serial {
		fmt.Println("def")
		return false
	}
	if ipv4 := q["ipv4"]; len(ipv4) > 0 {
		match := false
		for _, ip := range m.Spec.IPv4 {
			if ip == ipv4 {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}
	if ipv6 := q["ipv6"]; len(ipv6) > 0 {
		match := false
		for _, ip := range m.Spec.IPv6 {
			if ip == ipv6 {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}
	if product := q["product"]; len(product) > 0 && product != m.Spec.Product {
		return false
	}
	if datacenter := q["datacenter"]; len(datacenter) > 0 && datacenter != m.Spec.Datacenter {
		return false
	}
	if rack := q["rack"]; len(rack) > 0 && rack != fmt.Sprint(m.Spec.Rack) {
		return false
	}
	if role := q["role"]; len(role) > 0 && role != m.Spec.Role {
		return false
	}
	if bmc := q["bmc-type"]; len(bmc) > 0 && bmc != m.Spec.BMC.Type {
		return false
	}
	if state := q["state"]; len(state) > 0 && state != m.Status.State.String() {
		return false
	}

	return true
}

// Serial returns value of serial in the query
func (q Query) Serial() string { return q["serial"] }

// Product returns value of product in the query
func (q Query) Product() string { return q["product"] }

// Datacenter returns value of datacenter in the query
func (q Query) Datacenter() string { return q["datacenter"] }

// Rack returns value of rack in the query
func (q Query) Rack() string { return q["rack"] }

// Role returns value of role in the query
func (q Query) Role() string { return q["role"] }

// IPv4 returns value of ipv4 in the query
func (q Query) IPv4() string { return q["ipv4"] }

// IPv6 returns value of ipv6 in the query
func (q Query) IPv6() string { return q["ipv6"] }

// BMCType returns value of bmc-type in the query
func (q Query) BMCType() string { return q["bmc-type"] }

// State returns value of state the query
func (q Query) State() string { return q["state"] }

// IsEmpty returns true if query is empty or no values are presented
func (q Query) IsEmpty() bool {
	for _, v := range q {
		if len(v) > 0 {
			return false
		}
	}
	return true
}
