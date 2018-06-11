package client

import "testing"

func TestConstructIgnitionYAML(t *testing.T) {
	type args struct {
		baseDir string
		source  *ignitionSource
		ignMap  map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := constructIgnitionYAML(tt.args.baseDir, tt.args.source, tt.args.ignMap); (err != nil) != tt.wantErr {
				t.Errorf("constructIgnitionYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
