//go:generate gorunpkg github.com/99designs/gqlgen

package gql

import (
	context "context"
)

type Resolver struct{}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Machine(ctx context.Context, serial string) (Machine, error) {
	panic("not implemented")
}
func (r *queryResolver) SearchMachines(ctx context.Context, having *MachineParams, notHaving *MachineParams) ([]Machine, error) {
	panic("not implemented")
}
