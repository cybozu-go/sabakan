package sabakan

import "testing"

func TestValidatePostParams(t *testing.T) {
	valid := sabakanCrypt{"disk-a", "fooo"}
	if err := validatePostParams(valid); err != nil {
		t.Fatal("validator should return nil when the args are valid.")
	}
	invalid1 := sabakanCrypt{"", "fooo"}
	if err := validatePostParams(invalid1); err == nil {
		t.Fatal("validator should return error when the args are valid.")
	}
	invalid2 := sabakanCrypt{"disk-a", ""}
	if err := validatePostParams(invalid2); err == nil {
		t.Fatal("validator should return error when the args are valid.")
	}
}
