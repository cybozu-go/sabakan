package sabakan

import (
	"encoding/json"
	"testing"
)

func TestIsValidRole(t *testing.T) {
	t.Parallel()

	roles := []string{"ss", "cs", "Kube-worker1", "kube_master", "k8s.node"}
	for _, r := range roles {
		if !IsValidRole(r) {
			t.Error("validator should return true:", r)
		}
	}
	roles = []string{"s/s", "s s", "", "[kubernetes]api"}
	for _, r := range roles {
		if IsValidRole(r) {
			t.Error("validator should return false:", r)
		}
	}
}

func TestIsValidBmcType(t *testing.T) {
	t.Parallel()
	validTypes := []string{"validtype1", "valid_type2", "valid/type-3"}
	for _, ty := range validTypes {
		if !IsValidBmcType(ty) {
			t.Error("validator should return true:", ty)
		}
	}
	invalidTypes := []string{"invalid\\type 1", "%invalid+type#2", "^invalid$type?"}
	for _, ty := range invalidTypes {
		if IsValidBmcType(ty) {
			t.Error("validator should return false:", ty)
		}
	}
}

func TestMachine(t *testing.T) {
	t.Parallel()

	spec := MachineSpec{
		Serial: "abc",
		Rack:   3,
		Role:   "boot",
		BMC:    MachineBMC{Type: "IPMI-2.0"},
	}

	m := NewMachine(spec)
	if m.Status.State != StateHealthy {
		t.Error(`m.Status.State != StateHealthy`)
	}
	if m.Status.Timestamp.IsZero() {
		t.Error(`m.Status.Timestamp.IsZero()`)
	}

	j, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	m2 := new(Machine)
	err = json.Unmarshal(j, m2)
	if err != nil {
		t.Fatal(err)
	}
	if m2.Status.State != StateHealthy {
		t.Error(`m.Status.State != StateHealthy`)
	}
}
