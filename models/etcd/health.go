package etcd

import (
	"context"
	"time"

	"github.com/cybozu-go/sabakan"
)

func (d *driver) getHealth(ctx context.Context) (sabakan.HealthStatus, error) {
	var hs sabakan.HealthStatus
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	_, err := d.client.Get(ctx, "health")

	if err != nil {
		hs.Health = "unhealty"
		return hs, err
	}

	hs.Health = "healty"
	return hs, err

}

type healthDriver struct {
	*driver
}

func (d healthDriver) GetHealth(ctx context.Context) (sabakan.HealthStatus, error) {
	return d.getHealth(ctx)
}
