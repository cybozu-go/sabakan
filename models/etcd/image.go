package etcd

import (
	"context"
	"encoding/json"
	"io"
	"path"
	"path/filepath"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan"
)

func (d *driver) getImageDir(os string) ImageDir {
	return ImageDir{
		Dir: filepath.Join(d.imageDir, os),
	}
}

func (d *driver) imageGetIndexWithRev(ctx context.Context, os string) (sabakan.ImageIndex, int64, error) {
	key := path.Join(d.prefix, KeyImages, os)
	resp, err := d.client.Get(ctx, key)
	if err != nil {
		return nil, 0, err
	}
	if resp.Count == 0 {
		return sabakan.ImageIndex{}, 0, nil
	}

	var ret sabakan.ImageIndex
	err = json.Unmarshal(resp.Kvs[0].Value, &ret)
	if err != nil {
		return nil, 0, err
	}
	dir := d.getImageDir(os)
	for _, img := range ret {
		img.Exists = dir.Exists(img.ID)
	}
	return ret, resp.Kvs[0].ModRevision, nil
}

func (d *driver) imageGetDeletedWithRev(ctx context.Context, os string) ([]string, int64, error) {
	key := path.Join(d.prefix, KeyImages, os, "deleted")
	resp, err := d.client.Get(ctx, key)
	if err != nil {
		return nil, 0, err
	}
	if resp.Count == 0 {
		return nil, 0, nil
	}

	var ret []string
	err = json.Unmarshal(resp.Kvs[0].Value, &ret)
	if err != nil {
		return nil, 0, err
	}
	return ret, resp.Kvs[0].ModRevision, nil
}

func (d *driver) imageGetIndex(ctx context.Context, os string) (sabakan.ImageIndex, error) {
	index, _, err := d.imageGetIndexWithRev(ctx, os)
	return index, err
}

func (d *driver) imageCASIndex(ctx context.Context, os string,
	index sabakan.ImageIndex, indexRev int64,
	deleted []string, delRev int64) (*clientv3.TxnResponse, error) {

	indexKey := path.Join(d.prefix, KeyImages, os)
	deletedKey := path.Join(d.prefix, KeyImages, os, "deleted")

	indexJSON, err := json.Marshal(index)
	if err != nil {
		return nil, err
	}
	deletedJSON, err := json.Marshal(deleted)
	if err != nil {
		return nil, err
	}

	return d.client.Txn(ctx).
		If(
			clientv3.Compare(clientv3.ModRevision(indexKey), "=", indexRev),
			clientv3.Compare(clientv3.ModRevision(deletedKey), "=", delRev),
		).
		Then(
			clientv3.OpPut(indexKey, string(indexJSON)),
			clientv3.OpPut(deletedKey, string(deletedJSON)),
		).
		Commit()
}

func (d *driver) imageUpload(ctx context.Context, os, id string, r io.Reader) error {
RETRY:
	index, indexRev, err := d.imageGetIndexWithRev(ctx, os)
	if err != nil {
		return err
	}
	deleted, delRev, err := d.imageGetDeletedWithRev(ctx, os)
	if err != nil {
		return err
	}

	img := index.Find(id)
	if img != nil {
		return sabakan.ErrConflicted
	}

	dir := d.getImageDir(os)
	err = dir.Extract(r, id, []string{sabakan.ImageKernelFilename, sabakan.ImageInitrdFilename})
	if err != nil {
		return err
	}

	index, dels := index.Append(&sabakan.Image{
		ID:   id,
		Date: time.Now().UTC(),
		URLs: []string{d.myURL("/api/v1/images", os, id)},
	})
	deleted = append(deleted, dels...)
	if len(deleted) > MaxDeleted {
		deleted = deleted[len(deleted)-MaxDeleted:]
	}

	resp, err := d.imageCASIndex(ctx, os, index, indexRev, deleted, delRev)
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		goto RETRY
	}

	return nil
}

func (d *driver) imageDownload(ctx context.Context, os, id string, out io.Writer) error {
	index, err := d.imageGetIndex(ctx, os)
	if err != nil {
		return err
	}

	img := index.Find(id)
	if img == nil {
		return sabakan.ErrNotFound
	}

	dir := d.getImageDir(os)
	if !dir.Exists(id) {
		return sabakan.ErrNotFound
	}

	err = dir.Download(out, id)
	if err != nil {
		log.Error("imageDownload failed", map[string]interface{}{
			"os":        os,
			"id":        id,
			log.FnError: err.Error(),
		})
	}
	return err
}

func (d *driver) imageDelete(ctx context.Context, os, id string) error {
RETRY:
	index, indexRev, err := d.imageGetIndexWithRev(ctx, os)
	if err != nil {
		return err
	}
	deleted, delRev, err := d.imageGetDeletedWithRev(ctx, os)
	if err != nil {
		return err
	}

	if len(index) == 0 {
		return sabakan.ErrNotFound
	}

	newIndex := make(sabakan.ImageIndex, 0, len(index))
	for _, img := range index {
		if img.ID == id {
			continue
		}
		newIndex = append(newIndex, img)
	}
	if len(newIndex) == len(index) {
		return sabakan.ErrNotFound
	}

	deleted = append(deleted, id)
	if len(deleted) > MaxDeleted {
		deleted = deleted[len(deleted)-MaxDeleted:]
	}

	resp, err := d.imageCASIndex(ctx, os, newIndex, indexRev, deleted, delRev)
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		goto RETRY
	}

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
