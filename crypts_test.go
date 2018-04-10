package sabakan

import "testing"

func Test_validatePostParams(t *testing.T) {
	validEntity := cryptEntity{"disk-a", "fooo"}
	if err := validatePostParams(validEntity); err != nil {
		t.Fatal("validator should return nil when the args are valid.")
	}
	invalidEntity1 := cryptEntity{"", "fooo"}
	if err := validatePostParams(invalidEntity1); err == nil {
		t.Fatal("validator should return error when the args are valid.")
	}
	invalidEntity2 := cryptEntity{"disk-a", ""}
	if err := validatePostParams(invalidEntity2); err == nil {
		t.Fatal("validator should return error when the args are valid.")
	}
}
