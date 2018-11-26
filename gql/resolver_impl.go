//go:generate gorunpkg github.com/99designs/gqlgen

package gql

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan"
)

// Resolver implements ResolverRoot.
type Resolver struct {
	Model sabakan.Model
}

// BMC implements ResolverRoot.
func (r *Resolver) BMC() BMCResolver {
	return &bMCResolver{r}
}

// Machine implements ResolverRoot.
func (r *Resolver) Machine() MachineResolver {
	return &machineResolver{r}
}

// MachineStatus implements ResolverRoot.
func (r *Resolver) MachineStatus() MachineStatusResolver {
	return &machineStatusResolver{r}
}

// Query implements ResolverRoot.
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type bMCResolver struct{ *Resolver }

func (r *bMCResolver) BmcType(ctx context.Context, obj *sabakan.MachineBMC) (string, error) {
	return obj.Type, nil
}
func (r *bMCResolver) Ipv4(ctx context.Context, obj *sabakan.MachineBMC) (IPAddress, error) {
	return IPAddress(net.ParseIP(obj.IPv4)), nil
}

type machineResolver struct{ *Resolver }

func (r *machineResolver) Serial(ctx context.Context, obj *sabakan.Machine) (string, error) {
	return obj.Spec.Serial, nil
}
func (r *machineResolver) Labels(ctx context.Context, obj *sabakan.Machine) ([]Label, error) {
	labels := make([]Label, 0, len(obj.Spec.Labels))
	for k, v := range obj.Spec.Labels {
		labels = append(labels, Label{Name: k, Value: v})
	}
	return labels, nil
}
func (r *machineResolver) Rack(ctx context.Context, obj *sabakan.Machine) (int, error) {
	return int(obj.Spec.Rack), nil
}
func (r *machineResolver) IndexInRack(ctx context.Context, obj *sabakan.Machine) (int, error) {
	return int(obj.Spec.IndexInRack), nil
}
func (r *machineResolver) Role(ctx context.Context, obj *sabakan.Machine) (string, error) {
	return obj.Spec.Role, nil
}
func (r *machineResolver) Ipv4(ctx context.Context, obj *sabakan.Machine) ([]IPAddress, error) {
	addresses := make([]IPAddress, len(obj.Spec.IPv4))
	for i, a := range obj.Spec.IPv4 {
		addresses[i] = IPAddress(net.ParseIP(a))
	}
	return addresses, nil
}
func (r *machineResolver) RegisterDate(ctx context.Context, obj *sabakan.Machine) (DateTime, error) {
	return DateTime(obj.Spec.RegisterDate), nil
}
func (r *machineResolver) RetireDate(ctx context.Context, obj *sabakan.Machine) (DateTime, error) {
	return DateTime(obj.Spec.RetireDate), nil
}
func (r *machineResolver) Bmc(ctx context.Context, obj *sabakan.Machine) (sabakan.MachineBMC, error) {
	return obj.Spec.BMC, nil
}

type machineStatusResolver struct{ *Resolver }

func (r *machineStatusResolver) State(ctx context.Context, obj *sabakan.MachineStatus) (MachineState, error) {
	switch obj.State {
	case sabakan.StateUninitialized:
		return MachineStateUninitialized, nil
	case sabakan.StateHealthy:
		return MachineStateHealthy, nil
	case sabakan.StateUnhealthy:
		return MachineStateUnhealthy, nil
	case sabakan.StateUnreachable:
		return MachineStateUnreachable, nil
	case sabakan.StateUpdating:
		return MachineStateUpdating, nil
	case sabakan.StateRetiring:
		return MachineStateRetiring, nil
	case sabakan.StateRetired:
		return MachineStateRetired, nil
	default:
		return "", fmt.Errorf("unknown state:%s", obj.State.String())
	}
}
func (r *machineStatusResolver) Timestamp(ctx context.Context, obj *sabakan.MachineStatus) (DateTime, error) {
	return DateTime(obj.Timestamp), nil
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Machine(ctx context.Context, serial string) (sabakan.Machine, error) {
	now := time.Now()
	machine, err := r.Model.Machine.Get(ctx, serial)
	if err != nil {
		return sabakan.Machine{}, err
	}
	machine.Status.Duration = now.Sub(machine.Status.Timestamp).Seconds()
	return *machine, nil
}
func (r *queryResolver) SearchMachines(ctx context.Context, having, notHaving *MachineParams) ([]sabakan.Machine, error) {
	now := time.Now()
	machines, err := r.Model.Machine.Query(ctx, sabakan.Query{})
	if err != nil {
		return nil, err
	}
	var filtered []sabakan.Machine
	for _, m := range machines {
		m.Status.Duration = now.Sub(m.Status.Timestamp).Seconds()
		if matchMachine(m, having) && !(matchMachine(m, notHaving)) {
			filtered = append(filtered, *m)
		}
	}
	return filtered, nil
}

func matchMachine(machine *sabakan.Machine, having *MachineParams) bool {
	if !containsAllInputLabels(having.Labels, machine.Spec.Labels) {
		return false
	}
	if !containsRack(having.Racks, int(machine.Spec.Rack)) {
		return false
	}
	if !containsRole(having.Roles, machine.Spec.Role) {
		return false
	}
	if !containsStates(having.States, machine.Status.State) {
		return false
	}
	if isOlderThan(*having.MinDaysBeforeRetire, machine.Status.Duration) {
		return false
	}

	return true
}

func isOlderThan(minDaysBeforeRetire int, currentDuration float64) bool {
	dur, err := time.ParseDuration(fmt.Sprintf("%dh", 24*minDaysBeforeRetire))
	if err != nil {
		log.Error("failed to parse duration", map[string]interface{}{
			log.FnError: err.Error(),
		})
		return false
	}
	return dur.Seconds() > currentDuration
}

func containsStates(states []MachineState, target sabakan.MachineState) bool {
	for _, state := range states {
		if state.String() == strings.ToUpper(target.String()) {
			return true
		}
	}
	return false
}

func containsRole(roles []string, target string) bool {
	for _, role := range roles {
		if role == target {
			return true
		}
	}
	return false
}

func containsRack(racks []int, target int) bool {
	for _, rack := range racks {
		if rack == target {
			return true
		}
	}
	return false
}

func containsAllInputLabels(labelInputs []LabelInput, labels map[string]string) bool {
	for _, input := range labelInputs {
		if !containsLabel(input, labels) {
			return false
		}
	}
	return true
}

func containsLabel(input LabelInput, labels map[string]string) bool {
	for k, v := range labels {
		if (input.Name == k) && (input.Value == v) {
			return true
		}
	}
	return false
}
