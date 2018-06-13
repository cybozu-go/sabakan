package etcd

import (
	"context"

	"github.com/coreos/etcd/clientv3"
)

func (d *driver) initAssets(ctx context.Context, rev int64) error {
	return nil
}

func (d *driver) handleAssetEvent(ctx context.Context, ev *clientv3.Event) error {
	return nil
}

func (d *driver) startAssetUpdater(ctx context.Context, ch <-chan EventPool) error {
	for ep := range ch {
		for _, ev := range ep.Events {
			err := d.handleAssetEvent(ctx, ev)
			if err != nil {
				return err
			}
		}

		// checkpoint
		d.saveLastRev(ep.Rev)
	}
	return nil
}
