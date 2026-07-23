package sabakan

import (
	"fmt"
	"slices"
	"strings"
)

// Query is an URL query
type Query map[string]string

// Match returns true if all non-empty fields matches Machine
func (q Query) Match(m *Machine) (bool, error) {
	if serial := q["serial"]; len(serial) > 0 {
		if !slices.Contains(strings.Split(serial, ","), m.Spec.Serial) {
			return false, nil
		}
	}
	if ipv4 := q["ipv4"]; len(ipv4) > 0 {
		match := false
		for ipv4address := range strings.SplitSeq(ipv4, ",") {
			if slices.Contains(m.Spec.IPv4, ipv4address) {
				match = true
				break
			}
		}
		if !match {
			return false, nil
		}
	}
	if ipv6 := q["ipv6"]; len(ipv6) > 0 {
		match := false
		for ipv6address := range strings.SplitSeq(ipv6, ",") {
			if slices.Contains(m.Spec.IPv6, ipv6address) {
				match = true
				break
			}
		}
		if !match {
			return false, nil
		}
	}
	if labels := q["labels"]; len(labels) > 0 {
		for query := range strings.SplitSeq(labels, ",") {
			kv := strings.SplitN(query, "=", 2)
			if len(kv) != 2 {
				return false, fmt.Errorf("invalid query in labels: %s", query)
			}
			queryKey := kv[0]
			queryValue := kv[1]
			if value, exists := m.Spec.Labels[queryKey]; exists {
				if value != queryValue {
					return false, nil
				}
			} else {
				return false, nil
			}
		}
	}
	if rack := q["rack"]; len(rack) > 0 {
		if !slices.Contains(strings.Split(rack, ","), fmt.Sprint(m.Spec.Rack)) {
			return false, nil
		}
	}
	if role := q["role"]; len(role) > 0 {
		if !slices.Contains(strings.Split(role, ","), m.Spec.Role) {
			return false, nil
		}
	}
	if bmc := q["bmc-type"]; len(bmc) > 0 {
		if !slices.Contains(strings.Split(bmc, ","), m.Spec.BMC.Type) {
			return false, nil
		}
	}
	if state := q["state"]; len(state) > 0 {
		if !slices.Contains(strings.Split(state, ","), fmt.Sprint(m.Status.State)) {
			return false, nil
		}
	}
	if withoutSerial := q["without-serial"]; len(withoutSerial) > 0 {
		if slices.Contains(strings.Split(withoutSerial, ","), fmt.Sprint(m.Spec.Serial)) {
			return false, nil
		}
	}
	if withoutIPv4 := q["without-ipv4"]; len(withoutIPv4) > 0 {
		for wIPv4 := range strings.SplitSeq(withoutIPv4, ",") {
			if slices.Contains(m.Spec.IPv4, wIPv4) {
				return false, nil
			}
		}
	}
	if withoutIPv6 := q["without-ipv6"]; len(withoutIPv6) > 0 {
		for wIPv6 := range strings.SplitSeq(withoutIPv6, ",") {
			if slices.Contains(m.Spec.IPv6, wIPv6) {
				return false, nil
			}
		}
	}
	if withoutLabels := q["without-labels"]; len(withoutLabels) > 0 {
		excluded := true
		for query := range strings.SplitSeq(withoutLabels, ",") {
			kv := strings.SplitN(query, "=", 2)
			if len(kv) != 2 {
				return false, fmt.Errorf("invalid query in without-labels: %s", query)
			}
			queryKey := kv[0]
			queryValue := kv[1]
			if value, exists := m.Spec.Labels[queryKey]; exists {
				if value != queryValue {
					excluded = false
					break
				}
			} else {
				excluded = false
				break
			}
		}
		if excluded {
			return false, nil
		}
	}
	if withoutRack := q["without-rack"]; len(withoutRack) > 0 {
		if slices.Contains(strings.Split(withoutRack, ","), fmt.Sprint(m.Spec.Rack)) {
			return false, nil
		}
	}
	if withoutRole := q["without-role"]; len(withoutRole) > 0 {
		if slices.Contains(strings.Split(withoutRole, ","), fmt.Sprint(m.Spec.Role)) {
			return false, nil
		}
	}
	if withoutBmc := q["without-bmc-type"]; len(withoutBmc) > 0 {
		if slices.Contains(strings.Split(withoutBmc, ","), fmt.Sprint(m.Spec.BMC.Type)) {
			return false, nil
		}
	}
	if withoutState := q["without-state"]; len(withoutState) > 0 {
		if slices.Contains(strings.Split(withoutState, ","), fmt.Sprint(m.Status.State)) {
			return false, nil
		}
	}

	return true, nil
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
func (q Query) HasOnlyWithout() bool {
	for k, v := range q {
		if !strings.HasPrefix(k, "without") && len(v) > 0 {
			return false
		}
	}
	return true
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
