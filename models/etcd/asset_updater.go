package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/coreos/etcd/clientv3"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan"
)

func (d *driver) addAssetURL(ctx context.Context, asset *sabakan.Asset) error {
RETRY:
	a, rev, err := d.assetGetInfoWithRev(ctx, asset.Name)
	if err != nil {
		if err == sabakan.ErrNotFound {
			// asset has been deleted
			return nil
		}
		return err
	}

	// asset has been replaced
	if a.ID != asset.ID {
		return nil
	}

	a.URLs = append(a.URLs, d.myURL("/api/v1/assets", a.Name))
	key := KeyAssets + a.Name

	j, err := json.Marshal(a)
	if err != nil {
		return err
	}

	resp, err := d.client.Txn(ctx).
		If(clientv3.Compare(clientv3.ModRevision(key), "=", rev)).
		Then(clientv3.OpPut(key, string(j))).
		Commit()
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		goto RETRY
	}

	return nil
}

func (d *driver) downloadAsset(ctx context.Context, asset *sabakan.Asset) error {
	urls := make([]string, len(asset.URLs))
	for i, v := range rand.Perm(len(urls)) {
		urls[v] = asset.URLs[i]
	}

	var resp *http.Response
	var err error

	for _, u := range urls {
		resp, err = d.pullURL(ctx, u)
		if err == nil {
			break
		}
	}

	if err != nil {
		log.Error("failed to pull an asset", map[string]interface{}{
			"name": asset.Name,
			"id":   asset.ID,
			"urls": asset.URLs,
		})
		return err
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	id, err := strconv.Atoi(resp.Header.Get("X-Sabakan-Asset-ID"))
	if err != nil {
		log.Error("invalid asset ID", map[string]interface{}{
			log.FnError: err,
			"id":        resp.Header.Get("X-Sabakan-Asset-ID"),
		})
		return errors.New("invalid asset ID")
	}
	csum := resp.Header.Get("X-Sabakan-Asset-SHA256")

	if asset.ID != id {
		// the asset has been replaced with newer one
		log.Info("received newer asset", map[string]interface{}{
			"expected": asset.ID,
			"received": id,
			"name":     asset.Name,
		})
		return nil
	}

	_, err = d.getAssetDir().Save(id, resp.Body, csum)
	if err != nil {
		return err
	}

	if len(asset.URLs) < maxAssetURLs {
		err = d.addAssetURL(ctx, asset)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *driver) initAssets(ctx context.Context, rev int64) error {
	resp, err := d.client.Get(ctx, KeyAssets,
		clientv3.WithPrefix(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
		clientv3.WithLimit(assetPageSize),
		clientv3.WithRev(rev))
	if err != nil {
		return err
	}

	dir := d.getAssetDir()
	ids := make(map[int]bool)
	kvs := resp.Kvs

REDO:
	for _, kv := range kvs {
		asset := new(sabakan.Asset)
		err = json.Unmarshal(kv.Value, asset)
		if err != nil {
			return err
		}
		ids[asset.ID] = true
		if dir.Exists(asset.ID) {
			continue
		}
		err = d.downloadAsset(ctx, asset)
		if err != nil {
			return err
		}
	}

	if resp.More {
		resp, err = d.client.Get(ctx, string(resp.Kvs[len(resp.Kvs)-1].Key),
			clientv3.WithRange(clientv3.GetPrefixRangeEnd(KeyAssets)),
			clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
			clientv3.WithLimit(assetPageSize),
			clientv3.WithRev(rev))
		if err != nil {
			return err
		}

		// ignore the first key
		kvs = resp.Kvs[1:]
		goto REDO
	}

	dir.GC(ids)
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
