package gql

import (
	"slices"
	"time"

	"github.com/cybozu-go/sabakan/v3"
	"github.com/cybozu-go/sabakan/v3/gql/graph/model"
)

// MatchMachine tests if a machine matches the given conditions
func MatchMachine(m *sabakan.Machine, h, nh *model.MachineParams, now time.Time) bool {
	if !containsAllLabels(h, m.Spec.Labels) {
		return false
	}
	if containsAnyLabel(nh, m.Spec.Labels) {
		return false
	}

	if !containsRack(h, int(m.Spec.Rack), true) {
		return false
	}
	if containsRack(nh, int(m.Spec.Rack), false) {
		return false
	}

	if !containsRole(h, m.Spec.Role, true) {
		return false
	}
	if containsRole(nh, m.Spec.Role, false) {
		return false
	}

	if !containsState(h, m.Status.State, true) {
		return false
	}
	if containsState(nh, m.Status.State, false) {
		return false
	}

	days := int(m.Spec.RetireDate.Sub(now).Hours() / 24)
	if h != nil && h.MinDaysBeforeRetire != nil {
		if *h.MinDaysBeforeRetire > days {
			return false
		}
	}
	if nh != nil && nh.MinDaysBeforeRetire != nil {
		if *nh.MinDaysBeforeRetire <= days {
			return false
		}
	}

	return true
}

func containsAllLabels(h *model.MachineParams, labels map[string]string) bool {
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

func containsAnyLabel(h *model.MachineParams, labels map[string]string) bool {
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

func containsRack(h *model.MachineParams, target int, base bool) bool {
	if h == nil || len(h.Racks) == 0 {
		return base
	}
	return slices.Contains(h.Racks, target)
}

func containsRole(h *model.MachineParams, target string, base bool) bool {
	if h == nil || len(h.Roles) == 0 {
		return base
	}
	return slices.Contains(h.Roles, target)
}

func containsState(h *model.MachineParams, target sabakan.MachineState, base bool) bool {
	if h == nil || len(h.States) == 0 {
		return base
	}

	return slices.Contains(h.States, target)
}
