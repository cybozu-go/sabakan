package etcd

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/cybozu-go/sabakan/v2"
)

func (d *driver) getAssetDir() AssetDir {
	return AssetDir{
		Dir: filepath.Join(d.dataDir, "assets"),
	}
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

func (d *driver) assetGetInfoWithRev(ctx context.Context, name string) (*sabakan.Asset, int64, error) {
	key := KeyAssets + name
	resp, err := d.client.Get(ctx, key)
	if err != nil {
		return nil, 0, err
	}

	if resp.Count == 0 {
		return nil, 0, sabakan.ErrNotFound
	}

	a := new(sabakan.Asset)
	err = json.Unmarshal(resp.Kvs[0].Value, a)
	if err != nil {
		return nil, 0, err
	}

	return a, resp.Kvs[0].ModRevision, nil
}

func (d *driver) assetGetInfo(ctx context.Context, name string) (*sabakan.Asset, error) {
	a, _, err := d.assetGetInfoWithRev(ctx, name)
	if a != nil {
		a.Exists = d.getAssetDir().Exists(a.ID)
	}
	return a, err
}

func (d *driver) assetPut(ctx context.Context, name, contentType string,
	csum []byte, options map[string]string, r io.Reader) (*sabakan.AssetStatus, error) {
	id, err := d.assetNewID(ctx)
	if err != nil {
		return nil, err
	}

	dir := d.getAssetDir()
	hsum, err := dir.Save(id, r, csum)
	if err != nil {
		return nil, err
	}

	hsumString := hex.EncodeToString(hsum)
	a := &sabakan.Asset{
		Name:        name,
		ID:          id,
		ContentType: contentType,
		Date:        time.Now().UTC(),
		Sha256:      hsumString,
		Options:     options,
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
		dir.Remove(id)
		return nil, sabakan.ErrConflicted
	}

	d.addLog(ctx, time.Now(), tresp.Header.Revision, sabakan.AuditAssets,
		name, "put", "new checksum: "+hsumString)

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

	dir := d.getAssetDir()

	if dir.Exists(a.ID) {
		g, err := os.Open(dir.Path(a.ID))
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

	d.addLog(ctx, time.Now(), resp.Header.Revision, sabakan.AuditAssets, name, "delete", "")

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

func (d assetDriver) Put(ctx context.Context, name, contentType string,
	csum []byte, options map[string]string, r io.Reader) (*sabakan.AssetStatus, error) {
	return d.assetPut(ctx, name, contentType, csum, options, r)
}

func (d assetDriver) Get(ctx context.Context, name string, h sabakan.AssetHandler) error {
	return d.assetGet(ctx, name, h)
}

func (d assetDriver) Delete(ctx context.Context, name string) error {
	return d.assetDelete(ctx, name)
}
