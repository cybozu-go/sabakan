package cmd

import (
	"os"
	"testing"
)

func TestEnvOrDefault(t *testing.T) {
	tests := []struct {
		name  string
		value string
		set   bool
		want  string
	}{
		{name: "unset", want: "http://localhost:10080"},
		{name: "empty", value: "", set: true, want: "http://localhost:10080"},
		{name: "set", value: "http://10.69.0.3:10080", set: true, want: "http://10.69.0.3:10080"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Setenv restores the original value after the test, including for the unset case.
			t.Setenv("SABACTL_TEST_SERVER", tt.value)
			if !tt.set {
				os.Unsetenv("SABACTL_TEST_SERVER")
			}
			if got := envOrDefault("SABACTL_TEST_SERVER", "http://localhost:10080"); got != tt.want {
				t.Errorf("envOrDefault() = %q, want %q", got, tt.want)
			}
		})
	}
}
