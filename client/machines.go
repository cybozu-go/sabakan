package client

import (
	"context"
	"io/ioutil"
	"path"
	"strings"

	"github.com/cybozu-go/sabakan"
)

// MachinesGet get machine information from sabakan server
func MachinesGet(ctx context.Context, params map[string]string) ([]sabakan.Machine, *Status) {
	var machines []sabakan.Machine
	err := client.getJSON(ctx, "machines", params, &machines)
	if err != nil {
		return nil, err
	}
	return machines, nil
}

// MachinesCreate create machines information to sabakan server
func MachinesCreate(ctx context.Context, specs []*sabakan.MachineSpec) *Status {
	return client.sendRequestWithJSON(ctx, "POST", "machines", specs)
}

// MachinesRemove removes machine information from sabakan server
func MachinesRemove(ctx context.Context, serial string) *Status {
	return client.sendRequest(ctx, "DELETE", path.Join("machines", serial), nil)
}

// MachinesSetState set the state of the machine on sabakan server
func MachinesSetState(ctx context.Context, serial string, state string) *Status {
	r := strings.NewReader(state)
	return client.sendRequest(ctx, "PUT", "state/"+serial, r)
}

// MachinesGetState get the state of the machine from sabakan server
func MachinesGetState(ctx context.Context, serial string) (sabakan.MachineState, *Status) {
	req := client.NewRequest(ctx, "GET", "state/"+serial, nil)
	resp, status := client.Do(req)
	if status != nil {
		return "", status
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", ErrorStatus(err)
	}
	return sabakan.MachineState(data), nil
}
