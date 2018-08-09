package etcd

import (
	"context"
	"path"
	"regexp"

	"github.com/cybozu-go/sabakan"
)

func (d *driver) putParams(ctx context.Context, os string, params sabakan.KernelParams) error {
	r := regexp.MustCompile(`^([[:print:]])+$`)
	if !r.MatchString(string(params)) {
		return sabakan.ErrBadRequest
	}

	key := path.Join(KeyKernelParams, os)
	_, err := d.client.Put(ctx, key, string(params))
	if err != nil {
		return err
	}

	return nil
}

func (d *driver) getParams(ctx context.Context, os string) (sabakan.KernelParams, error) {
	key := path.Join(KeyKernelParams, os)
	resp, err := d.client.Get(ctx, key)
	if err != nil {
		return "", err
	}

	if resp.Count == 0 {
		return "", sabakan.ErrNotFound
	}

	v := resp.Kvs[0].Value
	return sabakan.KernelParams(v), nil
}

type kernelParamsDriver struct {
	*driver
}

func (d kernelParamsDriver) PutParams(ctx context.Context, os string, params sabakan.KernelParams) error {
	return d.putParams(ctx, os, params)
}

func (d kernelParamsDriver) GetParams(ctx context.Context, os string) (sabakan.KernelParams, error) {
	return d.getParams(ctx, os)
}
