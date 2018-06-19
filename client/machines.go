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
func MachinesCreate(ctx context.Context, specs []*sabakan.MachineSpec) *Status {
	return client.sendRequestWithJSON(ctx, "POST", "/machines", specs)
}

// MachinesRemove removes machine information from sabakan server
func MachinesRemove(ctx context.Context, serial string) *Status {
	return client.sendRequest(ctx, "DELETE", path.Join("/machines", serial))
}

// MachinesSetState set the state of the machine on sabakan server
func MachinesSetState(ctx context.Context, serial string, state string) *Status {
	return client.sendRequestWithBytes(ctx, "PUT", path.Join("/state", serial), []byte(state))
}

// MachinesSetState get the state of the machine from sabakan server
func MachinesGetState(ctx context.Context, serial string) (sabakan.MachineState, *Status) {
	state, err := client.getBytes(ctx, path.Join("/state", serial))
	if err != nil {
		return "", err
	}
	return sabakan.MachineState(state), nil
}
