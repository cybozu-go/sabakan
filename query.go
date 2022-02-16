package sabakan

import (
	"fmt"
	"net/url"
	"strings"
)

// Query is an URL query
type Query map[string]string

// Match returns true if all non-empty fields matches Machine
func (q Query) Match(m *Machine) bool {
	if serial := q["serial"]; len(serial) > 0 && serial != m.Spec.Serial {
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
	if labels := q["labels"]; len(labels) > 0 {
		// Split into each query
		rawQueries := strings.Split(labels, ",")
		for _, rawQuery := range rawQueries {
			rawQuery = strings.TrimSpace(rawQuery)
			query, err := url.ParseQuery(rawQuery)
			if err != nil {
				return false
			}
			for k, v := range query {
				if label, exists := m.Spec.Labels[k]; exists {
					if v[0] != label {
						return false
					}
				} else {
					return false
				}
			}
		}
	}
	if rack := q["rack"]; len(rack) > 0 {
		racks := strings.Split(rack, ",")
		match := false
		for _, rackname := range racks {
			if rackname == fmt.Sprint(m.Spec.Rack) {
				match = true
				break
			}
		}
		if !match {
			return false
		}
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
	if withoutRole := q["without-role"]; len(withoutRole) > 0 && withoutRole == m.Spec.Role {
		return false
	}
	if withoutBmc := q["without-bmc-type"]; len(withoutBmc) > 0 && withoutBmc == m.Spec.BMC.Type {
		return false
	}
	if withoutState := q["without-state"]; len(withoutState) > 0 && withoutState == m.Status.State.String() {
		return false
	}
	if withoutSerial := q["without-serial"]; len(withoutSerial) > 0 && withoutSerial == m.Spec.Serial {
		return false
	}
	if withoutIpv4 := q["without-ipv4"]; len(withoutIpv4) > 0 {
		for _, ip := range m.Spec.IPv4 {
			if ip == withoutIpv4 {
				return false
			}
		}
	}
	if withoutIpv6 := q["without-ipv6"]; len(withoutIpv6) > 0 {
		for _, ip := range m.Spec.IPv6 {
			if ip == withoutIpv6 {
				return false
			}
		}
	}
	if withoutLabels := q["without-labels"]; len(withoutLabels) > 0 {
		// Split into each query
		rawQueries := strings.Split(withoutLabels, ",")
		for _, rawQuery := range rawQueries {
			rawQuery = strings.TrimSpace(rawQuery)
			query, err := url.ParseQuery(rawQuery)
			if err != nil {
				return false
			}
			for k, v := range query {
				if label, exists := m.Spec.Labels[k]; exists {
					if v[0] == label {
						return false
					}
				} else {
					return false
				}
			}
		}
	}
	if withoutRack := q["without-rack"]; len(withoutRack) > 0 {
		withoutRacks := strings.Split(withoutRack, ",")
		for _, rackname := range withoutRacks {
			if rackname == fmt.Sprint(m.Spec.Rack) {
				return false
			}
		}
	}
	if withoutRole := q["without-role"]; len(withoutRole) > 0 && withoutRole == m.Spec.Role {
		return false
	}
	if withoutBmc := q["without-bmc-type"]; len(withoutBmc) > 0 && withoutBmc == m.Spec.BMC.Type {
		return false
	}
	if withoutState := q["without-state"]; len(withoutState) > 0 && withoutState == m.Status.State.String() {
		return false
	}

	return true
}

// Serial returns value of serial in the query
func (q Query) Serial() string { return q["serial"] }

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

// Labels return label's key and value combined with '='
func (q Query) Labels() []string {
	queries := strings.Split(q["labels"], ",")
	for idx, rawQuery := range queries {
		queries[idx] = strings.TrimSpace(rawQuery)
	}
	return queries
}

// IsEmpty returns true if query is empty or no values are presented
func (q Query) IsEmpty() bool {
	for _, v := range q {
		if len(v) > 0 {
			return false
		}
	}
	return true
}

// RemoveWithout returns query removed --without key
func (q Query) RemoveWithout() Query {
	removed := Query{}
	for k, v := range q {
		if strings.HasPrefix(k, "without") {
			removed[k] = ""
		} else {
			removed[k] = v
		}
	}
	return removed
}

// Valid returns true if query isn't conflicted
func (q Query) Valid() bool {
	hasWithoutSerial := q["without-serial"]
	if hasSerial := q["serial"]; len(hasSerial) > 0 && len(hasWithoutSerial) > 0 {
		return false
	}
	hasWithoutRack := q["without-rack"]
	if hasRack := q["rack"]; len(hasRack) > 0 && len(hasWithoutRack) > 0 {
		return false
	}
	hasWithoutRole := q["without-role"]
	if hasRole := q["role"]; len(hasRole) > 0 && len(hasWithoutRole) > 0 {
		return false
	}
	hasWithoutIPv4 := q["without-ipv4"]
	if hasIPv4 := q["ipv4"]; len(hasIPv4) > 0 && len(hasWithoutIPv4) > 0 {
		return false
	}
	hasWithoutIPv6 := q["without-ipv6"]
	if hasIPv6 := q["ipv6"]; len(hasIPv6) > 0 && len(hasWithoutIPv6) > 0 {
		return false
	}
	hasWithoutBMCType := q["without-bmc-type"]
	if hasBMCType := q["bmc-type"]; len(hasBMCType) > 0 && len(hasWithoutBMCType) > 0 {
		return false
	}
	hasWithoutState := q["without-state"]
	if hasState := q["state"]; len(hasState) > 0 && len(hasWithoutState) > 0 {
		return false
	}
	hasWithoutLabels := q["without-labels"]
	if hasLabels := q["labels"]; len(hasLabels) > 0 && len(hasWithoutLabels) > 0 {
		return false
	}
	return true
}
