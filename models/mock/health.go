package mock

import (
	"context"

	"github.com/cybozu-go/sabakan"
)

type healthDriver struct {
	health sabakan.HealthStatus
}

func newHealthDriver() *healthDriver {
	return &healthDriver{
		health: sabakan.HealthStatus{},
	}
}

func (d *healthDriver) GetHealth(ctx context.Context) (sabakan.HealthStatus, error) {
	var hs sabakan.HealthStatus
	hs.Health = "healthy"
	return hs, nil
}
