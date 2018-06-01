package sabakan

import "testing"

func TestIsValidRole(t *testing.T) {
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
