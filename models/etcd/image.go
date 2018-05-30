package etcd

import (
	"context"
	"io"
	"time"

	"github.com/cybozu-go/sabakan"
)

func (d *driver) imageGetIndex(ctx context.Context, os string) (sabakan.ImageIndex, error) {
	return nil, nil
}

func (d *driver) imageUpload(ctx context.Context, os, id string, r io.Reader) error {
	return nil
}

func (d *driver) imageDownload(ctx context.Context, os, id string, out io.Writer) error {
	return nil
}

func (d *driver) imageDelete(ctx context.Context, os, id string) error {
	return nil
}

func (d *driver) imageServeFile(ctx context.Context, os, filename string,
	f func(modtime time.Time, content io.ReadSeeker)) error {
	return nil
}

type imageDriver struct {
	*driver
}

func (d imageDriver) GetIndex(ctx context.Context, os string) (sabakan.ImageIndex, error) {
	return d.imageGetIndex(ctx, os)
}

func (d imageDriver) Upload(ctx context.Context, os, id string, r io.Reader) error {
	return d.imageUpload(ctx, os, id, r)
}

func (d imageDriver) Download(ctx context.Context, os, id string, out io.Writer) error {
	return d.imageDownload(ctx, os, id, out)
}

func (d imageDriver) Delete(ctx context.Context, os, id string) error {
	return d.imageDelete(ctx, os, id)
}

func (d imageDriver) ServeFile(ctx context.Context, os, filename string,
	f func(modtime time.Time, content io.ReadSeeker)) error {
	return d.imageServeFile(ctx, os, filename, f)
}
