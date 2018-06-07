package client

import (
	"context"
	"path"

	"github.com/cybozu-go/sabakan"
)

// MachinesGet get machine information from sabakan server
func MachinesGet(ctx context.Context, params map[string]string) ([]sabakan.Machine, *Status) {
	var machines []sabakan.Machine
	err := client.getJSON(ctx, "/machines", params, &machines)
	if err != nil {
		return nil, err
	}
	return machines, nil
}

// MachinesCreate create machines information to sabakan server
func MachinesCreate(ctx context.Context, machines []sabakan.Machine) *Status {
	return client.sendRequestWithJSON(ctx, "POST", "/machines", machines)
}

// MachinesRemove removes machine information from sabakan server
func MachinesRemove(ctx context.Context, serial string) *Status {
	return client.sendRequest(ctx, "DELETE", path.Join("/machines", serial))
}
