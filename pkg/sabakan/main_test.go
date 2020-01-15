package main

import (
	"net"
	"reflect"
	"testing"
)

func Test_parseAllowIPs(t *testing.T) {
	type args struct {
		ips []string
	}
	tests := []struct {
		name    string
		args    args
		want    []*net.IPNet
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAllowIPs(tt.args.ips)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseAllowIPs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseAllowIPs() got = %v, want %v", got, tt.want)
			}
		})
	}
}
