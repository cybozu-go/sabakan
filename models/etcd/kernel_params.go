package etcd

import (
	"context"
	"path"

	"github.com/cybozu-go/sabakan"
)

func (d *driver) putParams(ctx context.Context, os string, params string) error {
	key := path.Join(KeyKernelParams, os)
	_, err := d.client.Put(ctx, key, string(params))
	if err != nil {
		return err
	}

	return nil
}

func (d *driver) getParams(ctx context.Context, os string) (string, error) {
	key := path.Join(KeyKernelParams, os)
	resp, err := d.client.Get(ctx, key)
	if err != nil {
		return "", err
	}

	if resp.Count == 0 {
		return "", sabakan.ErrNotFound
	}

	v := resp.Kvs[0].Value
	return string(v), nil
}

type kernelParamsDriver struct {
	*driver
}

func (d kernelParamsDriver) PutParams(ctx context.Context, os string, params string) error {
	return d.putParams(ctx, os, params)
}

func (d kernelParamsDriver) GetParams(ctx context.Context, os string) (string, error) {
	return d.getParams(ctx, os)
}
