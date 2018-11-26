//go:generate gorunpkg github.com/99designs/gqlgen

package gql

import (
	"context"
	"fmt"
	"net"
	"sort"
	"time"

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
	if len(obj.Spec.Labels) == 0 {
		return []Label{}, nil
	}

	keys := make([]string, 0, len(obj.Spec.Labels))
	for k := range obj.Spec.Labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	labels := make([]Label, 0, len(obj.Spec.Labels))
	for _, k := range keys {
		labels = append(labels, Label{Name: k, Value: obj.Spec.Labels[k]})
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
		if matchMachine(m, having, notHaving, now) {
			filtered = append(filtered, *m)
		}
	}
	return filtered, nil
}
