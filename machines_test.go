package sabakan

import (
	"encoding/json"
	"reflect"
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

	validNames := []string{"valid_name1", "valid-n.ame2"}
	for _, vn := range validNames {
		if !IsValidLabelName(vn) {
			t.Error("validator should return true:", vn)
		}
	}
	invalidNames := []string{"^in;valid name\\1", "in$valid#name&2", "invalid@name=3", "invalid/name3"}
	for _, ivn := range invalidNames {
		if IsValidLabelName(ivn) {
			t.Error("validator should return false:", ivn)
		}
	}
}

func TestIsValidLabelValue(t *testing.T) {
	t.Parallel()

	validVals := []string{"validvalue1", "valid.value-_2"}
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

	labels := map[string]string{"datacenter": "Los Alamos", "product": "Cray-1"}
	for k, v := range labels {
		m.PutLabel(k, v)
	}
	if !reflect.DeepEqual(m.Spec.Labels, labels) {
		t.Error("m.Spec.Labels was not set correctly:", m.Spec.Labels)
	}
	m.PutLabel("datacenter", "Lawrence Livermore")
	if dc, ok := m.Spec.Labels["datacenter"]; !ok || dc != "Lawrence Livermore" {
		t.Error("m.Spec.Labels was not updated correctly:", m.Spec.Labels)
	}
	err = m.DeleteLabel("datacenter")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := m.Spec.Labels["datacenter"]; ok {
		t.Error("label in m.Spec.Labels was not deleted correctly:", m.Spec.Labels)
	}
}
