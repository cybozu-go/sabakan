package etcd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/cybozu-go/sabakan"
)

func (d *driver) assetDir() string {
	return filepath.Join(d.dataDir, "assets")
}

func (d *driver) assetPath(id int) string {
	return filepath.Join(d.assetDir(), strconv.Itoa(id))
}

func (d *driver) assetExists(id int) bool {
	_, err := os.Stat(d.assetPath(id))
	return err == nil
}

func (d *driver) assetNewID(ctx context.Context) (int, error) {
RETRY:
	resp, err := d.client.Get(ctx, KeyAssetsID)
	if err != nil {
		return 0, err
	}
	if resp.Count == 0 {
		_, err := d.client.Txn(ctx).
			If(clientv3util.KeyMissing(KeyAssetsID)).
			Then(clientv3.OpPut(KeyAssetsID, "0")).
			Commit()
		if err != nil {
			return 0, err
		}
		goto RETRY
	}

	rev := resp.Kvs[0].ModRevision
	id, err := strconv.Atoi(string(resp.Kvs[0].Value))
	if err != nil {
		return 0, err
	}
	id++
	value := strconv.Itoa(id)

	tresp, err := d.client.Txn(ctx).
		If(clientv3.Compare(clientv3.ModRevision(KeyAssetsID), "=", rev)).
		Then(clientv3.OpPut(KeyAssetsID, value)).
		Commit()
	if err != nil {
		return 0, err
	}
	if !tresp.Succeeded {
		goto RETRY
	}
	return id, nil
}

func (d *driver) assetGetIndex(ctx context.Context) ([]string, error) {
	resp, err := d.client.Get(ctx, KeyAssets,
		clientv3.WithPrefix(),
		clientv3.WithKeysOnly(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return nil, err
	}

	ret := make([]string, resp.Count)
	for i, kv := range resp.Kvs {
		ret[i] = string(kv.Key[len(KeyAssets):])
	}
	return ret, nil
}

func (d *driver) assetGetInfo(ctx context.Context, name string) (*sabakan.Asset, error) {
	key := KeyAssets + name
	resp, err := d.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if resp.Count == 0 {
		return nil, sabakan.ErrNotFound
	}

	a := new(sabakan.Asset)
	err = json.Unmarshal(resp.Kvs[0].Value, a)
	if err != nil {
		return nil, err
	}

	a.Exists = d.assetExists(a.ID)

	return a, nil
}

func (d *driver) assetPut(ctx context.Context, name, contentType string, r io.Reader) (*sabakan.AssetStatus, error) {
	id, err := d.assetNewID(ctx)
	if err != nil {
		return nil, err
	}

	dir := d.assetDir()
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}

	f, err := ioutil.TempFile(dir, ".tmp")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := sha256.New()
	w := io.MultiWriter(f, h)
	_, err = io.Copy(w, r)
	if err != nil {
		return nil, err
	}
	err = f.Sync()
	if err != nil {
		return nil, err
	}

	dest := d.assetPath(id)
	err = os.Rename(f.Name(), dest)
	if err != nil {
		return nil, err
	}

	a := &sabakan.Asset{
		Name:        name,
		ID:          id,
		ContentType: contentType,
		Date:        time.Now().UTC(),
		Sha256:      hex.EncodeToString(h.Sum(nil)),
		URLs:        []string{d.myURL("/api/v1/assets", name)},
	}
	data, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}

	key := KeyAssets + name
	resp, err := d.client.Get(ctx, key, clientv3.WithKeysOnly())
	if err != nil {
		return nil, err
	}

	ifop := clientv3util.KeyMissing(key)
	retStatus := http.StatusCreated

	if resp.Count != 0 {
		rev := resp.Kvs[0].ModRevision
		ifop = clientv3.Compare(clientv3.ModRevision(key), "=", rev)
		retStatus = http.StatusOK
	}

	tresp, err := d.client.Txn(ctx).
		If(ifop).Then(clientv3.OpPut(key, string(data))).Commit()
	if err != nil {
		return nil, err
	}
	if !tresp.Succeeded {
		os.Remove(dest)
		return nil, sabakan.ErrConflicted
	}

	return &sabakan.AssetStatus{
		Status: retStatus,
		ID:     id,
	}, nil
}

func (d *driver) assetGet(ctx context.Context, name string, h sabakan.AssetHandler) error {
	key := KeyAssets + name
	resp, err := d.client.Get(ctx, key)
	if err != nil {
		return err
	}

	if resp.Count == 0 {
		return sabakan.ErrNotFound
	}

	a := new(sabakan.Asset)
	err = json.Unmarshal(resp.Kvs[0].Value, a)
	if err != nil {
		return err
	}

	if d.assetExists(a.ID) {
		g, err := os.Open(d.assetPath(a.ID))
		if err != nil {
			return err
		}
		defer g.Close()

		h.ServeContent(a, g)
		return nil
	}

	h.Redirect(a.URLs[rand.Intn(len(a.URLs))])
	return nil
}

func (d *driver) assetDelete(ctx context.Context, name string) error {
	key := KeyAssets + name
	resp, err := d.client.Txn(ctx).
		If(clientv3util.KeyExists(key)).
		Then(clientv3.OpDelete(key)).
		Commit()

	if err != nil {
		return err
	}

	if !resp.Succeeded {
		return sabakan.ErrNotFound
	}
	return nil
}

type assetDriver struct {
	*driver
}

func (d assetDriver) GetIndex(ctx context.Context) ([]string, error) {
	return d.assetGetIndex(ctx)
}

func (d assetDriver) GetInfo(ctx context.Context, name string) (*sabakan.Asset, error) {
	return d.assetGetInfo(ctx, name)
}

func (d assetDriver) Put(ctx context.Context, name, contentType string, r io.Reader) (*sabakan.AssetStatus, error) {
	return d.assetPut(ctx, name, contentType, r)
}

func (d assetDriver) Get(ctx context.Context, name string, h sabakan.AssetHandler) error {
	return d.assetGet(ctx, name, h)
}

func (d assetDriver) Delete(ctx context.Context, name string) error {
	return d.assetDelete(ctx, name)
}
