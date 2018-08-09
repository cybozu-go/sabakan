package mock

import (
	"context"
	"sync"

	"github.com/cybozu-go/sabakan"
)

type kernelParamsDriver struct {
	mu           sync.Mutex
	kernelParams map[string]sabakan.KernelParams
}

func newKernelParamsDriver() *kernelParamsDriver {
	return &kernelParamsDriver{
		kernelParams: make(map[string]sabakan.KernelParams),
	}
}

func (d *kernelParamsDriver) PutParams(ctx context.Context, os string, params sabakan.KernelParams) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.kernelParams[os] = params
	return nil
}

func (d *kernelParamsDriver) GetParams(ctx context.Context, os string) (sabakan.KernelParams, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if val, ok := d.kernelParams[os]; ok {
		return val, nil
	}

	return "", sabakan.ErrNotFound
}
