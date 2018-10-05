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

func TestIsValidLabelName(t *testing.T) {
	t.Parallel()

	validNames := []string{"valid_name1", "valid-name2", "valid/name3"}
	for _, vn := range validNames {
		if !IsValidLabelName(vn) {
			t.Error("validator should return true:", vn)
		}
	}
	invalidNames := []string{"^in;valid name\\1", "in$valid#name&2", "invalid@name=3"}
	for _, ivn := range invalidNames {
		if IsValidLabelName(ivn) {
			t.Error("validator should return false:", ivn)
		}
	}
}

func TestIsValidLabelValue(t *testing.T) {
	t.Parallel()

	validVals := []string{"^valid value@1", "valid$value-=2", "%valid':value;3"}
	for _, vv := range validVals {
		if !IsValidLabelValue(vv) {
			t.Error("validator should return true:", vv)
		}
	}
	invalidVals := []string{`inválidvaluõ1`, `iñvalidvålue`}
	for _, ivv := range invalidVals {
		if IsValidLabelValue(ivv) {
			t.Error("validator should return false:", ivv)
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
	if m.Status.State != StateUninitialized {
		t.Error(`m.Status.State != StateUninitialized`)
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
	if m2.Status.State != StateUninitialized {
		t.Error(`m.Status.State != StateUninitialized`)
	}
}
