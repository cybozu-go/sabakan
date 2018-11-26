package gql

import (
	"fmt"
	"strings"
	"time"

	"github.com/cybozu-go/sabakan"
)

func matchMachine(m *sabakan.Machine, h, nh *MachineParams, now time.Time) bool {
	if !containsAllLabels(h, m.Spec.Labels) {
		fmt.Println("t1")
		return false
	}
	if containsAnyLabel(nh, m.Spec.Labels) {
		fmt.Println("t2")
		return false
	}

	if len(h.Racks) > 0 && !containsRack(h.Racks, int(m.Spec.Rack)) {
		fmt.Println("t3")
		return false
	}
	if containsRack(nh.Racks, int(m.Spec.Rack)) {
		fmt.Println("t4")
		return false
	}

	if len(h.Roles) > 0 && !containsRole(h.Roles, m.Spec.Role) {
		fmt.Println("t5")
		return false
	}
	if containsRole(nh.Roles, m.Spec.Role) {
		fmt.Println("t6")
		return false
	}

	if len(h.States) > 0 && !containsState(h.States, m.Status.State) {
		fmt.Println("t7")
		return false
	}
	if containsState(nh.States, m.Status.State) {
		fmt.Println("t8")
		return false
	}

	days := int(m.Spec.RetireDate.Sub(now).Hours() / 24)
	if h.MinDaysBeforeRetire != nil {
		if *h.MinDaysBeforeRetire > days {
			fmt.Println("t9")
			return false
		}
	}
	if nh.MinDaysBeforeRetire != nil {
		if *nh.MinDaysBeforeRetire <= days {
			fmt.Println("t10")
			return false
		}
	}

	return true
}

func containsAllLabels(h *MachineParams, labels map[string]string) bool {
	if h == nil {
		return true
	}
	for _, l := range h.Labels {
		v, ok := labels[l.Name]
		if !ok {
			return false
		}
		if v != l.Value {
			return false
		}
	}
	return true
}

func containsAnyLabel(h *MachineParams, labels map[string]string) bool {
	if h == nil {
		return false
	}
	for _, l := range h.Labels {
		v, ok := labels[l.Name]
		if !ok {
			continue
		}
		if v == l.Value {
			return true
		}
	}
	return false
}

func containsRack(racks []int, target int) bool {
	for _, rack := range racks {
		if rack == target {
			return true
		}
	}
	return false
}

func containsRole(roles []string, target string) bool {
	for _, role := range roles {
		if role == target {
			return true
		}
	}
	return false
}

func containsState(states []MachineState, target sabakan.MachineState) bool {
	for _, state := range states {
		if state.String() == strings.ToUpper(target.String()) {
			return true
		}
	}
	return false
}
