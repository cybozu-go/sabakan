package etcd

import (
	"context"
	"time"

	"go.etcd.io/etcd/etcdserver/api/v3rpc/rpctypes"
)

func (d *driver) getHealth(ctx context.Context) error {

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	_, err := d.client.Get(ctx, "health")

	if err == nil || err == rpctypes.ErrPermissionDenied {
		return nil
	}

	return err

}

type healthDriver struct {
	*driver
}

func (d healthDriver) GetHealth(ctx context.Context) error {
	return d.getHealth(ctx)
}
