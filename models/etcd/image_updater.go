package etcd

import (
	"context"
	"encoding/json"
	"path"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/cybozu-go/sabakan"
)

type updateData struct {
	index   sabakan.ImageIndex
	deleted []string
}

func (d *driver) updateImageForOS(ctx context.Context, os string, data updateData) error {
	return nil
}

func (d *driver) updateImage(ctx context.Context) error {
	key := path.Join(d.prefix, KeyImages) + "/"
	resp, err := d.client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	dataMap := make(map[string]updateData)
	for _, kv := range resp.Kvs {
		parts := strings.Split(string(kv.Key)[len(key):], "/")
		os := parts[0]

		if len(parts) == 1 {
			// this is the ImageIndex
			var index sabakan.ImageIndex
			err = json.Unmarshal(kv.Value, &index)
			if err != nil {
				return err
			}

			if data, ok := dataMap[os]; ok {
				data.index = index
			} else {
				dataMap[os] = updateData{index, nil}
			}
			continue
		}

		if len(parts) == 2 && parts[1] == "deleted" {
			var deleted []string
			err = json.Unmarshal(kv.Value, &deleted)
			if err != nil {
				return err
			}

			if data, ok := dataMap[os]; ok {
				data.deleted = deleted
			} else {
				dataMap[os] = updateData{nil, deleted}
			}
		}
	}

	for os, data := range dataMap {
		err = d.updateImageForOS(ctx, os, data)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *driver) startImageUpdater(ctx context.Context, ch <-chan struct{}) error {
	for {
		err := d.updateImage(ctx)
		if err != nil {
			return err
		}

		select {
		case <-ch:
		case <-ctx.Done():
			return nil
		}
	}
}
