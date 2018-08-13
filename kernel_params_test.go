package sabakan

import "testing"

func TestIsVlidKernelParams(t *testing.T) {
	t.Parallel()

	valids := []string{
		"console=ttyS0 coreos.autologin=ttyS0 glt",
		"",
	}
	for _, p := range valids {
		if !IsValidKernelParams(p) {
			t.Error(`!IsValidKernelParams(p): `, p)
		}
	}

	invalids := []string{
		"hello=workd\ngood=morning",
		"console=寿司",
	}
	for _, p := range invalids {
		if IsValidKernelParams(p) {
			t.Error(`IsValidKernelParams(p): `, p)
		}
	}
}
