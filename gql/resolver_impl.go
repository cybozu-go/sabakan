//go:generate gorunpkg github.com/99designs/gqlgen

package gql

import (
	context "context"

	sabakan "github.com/cybozu-go/sabakan"
)

// Resolver implements ResolverRoot.
type Resolver struct{}

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
	panic("not implemented")
}

type machineResolver struct{ *Resolver }

func (r *machineResolver) Serial(ctx context.Context, obj *sabakan.Machine) (string, error) {
	panic("not implemented")
}
func (r *machineResolver) Labels(ctx context.Context, obj *sabakan.Machine) ([]Label, error) {
	panic("not implemented")
}
func (r *machineResolver) Rack(ctx context.Context, obj *sabakan.Machine) (int, error) {
	panic("not implemented")
}
func (r *machineResolver) IndexInRack(ctx context.Context, obj *sabakan.Machine) (int, error) {
	panic("not implemented")
}
func (r *machineResolver) Role(ctx context.Context, obj *sabakan.Machine) (string, error) {
	panic("not implemented")
}
func (r *machineResolver) Ipv4(ctx context.Context, obj *sabakan.Machine) ([]IPAddress, error) {
	panic("not implemented")
}
func (r *machineResolver) RegisterDate(ctx context.Context, obj *sabakan.Machine) (DateTime, error) {
	panic("not implemented")
}
func (r *machineResolver) RetireDate(ctx context.Context, obj *sabakan.Machine) (DateTime, error) {
	panic("not implemented")
}
func (r *machineResolver) Bmc(ctx context.Context, obj *sabakan.Machine) (sabakan.MachineBMC, error) {
	panic("not implemented")
}

type machineStatusResolver struct{ *Resolver }

func (r *machineStatusResolver) State(ctx context.Context, obj *sabakan.MachineStatus) (MachineState, error) {
	panic("not implemented")
}
func (r *machineStatusResolver) Timestamp(ctx context.Context, obj *sabakan.MachineStatus) (DateTime, error) {
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Machine(ctx context.Context, serial string) (sabakan.Machine, error) {
	panic("not implemented")
}
func (r *queryResolver) SearchMachines(ctx context.Context, having *MachineParams, notHaving *MachineParams) ([]sabakan.Machine, error) {
	panic("not implemented")
}
