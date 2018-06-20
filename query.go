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

func (q Query) Serial() string     { return q["serial"] }
func (q Query) Product() string    { return q["product"] }
func (q Query) Datacenter() string { return q["datacenter"] }
func (q Query) Rack() string       { return q["rack"] }
func (q Query) Role() string       { return q["role"] }
func (q Query) IPv4() string       { return q["ipv4"] }
func (q Query) IPv6() string       { return q["ipv6"] }
func (q Query) BMCType() string    { return q["bmc-type"] }
func (q Query) State() string      { return q["state"] }

// IsEmpty returns true if query is empty
func (q Query) IsEmpty() bool {
	keys := map[string]struct{}{
		"serial":     struct{}{},
		"product":    struct{}{},
		"datacenter": struct{}{},
		"rack":       struct{}{},
		"role":       struct{}{},
		"ipv4":       struct{}{},
		"ipv6":       struct{}{},
		"bmc-type":   struct{}{},
		"state":      struct{}{},
	}

	for k, v := range q {
		if _, ok := keys[k]; ok && len(v) > 0 {
			return false
		}
	}
	return true
}
