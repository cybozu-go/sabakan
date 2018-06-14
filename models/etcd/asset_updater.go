package etcd

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan"
)

func decodeAsset(data []byte) (*sabakan.Asset, error) {
	asset := new(sabakan.Asset)
	err := json.Unmarshal(data, asset)
	if err != nil {
		return nil, err
	}
	return asset, err
}

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
		log.Error("asset: failed to pull", map[string]interface{}{
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
	csum, err := hex.DecodeString(resp.Header.Get("X-Sabakan-Asset-SHA256"))
	if err != nil {
		return err
	}

	if asset.ID != id {
		// the asset has been replaced with newer one
		log.Info("asset: replaced during pull", map[string]interface{}{
			"requested": asset.ID,
			"received":  id,
			"name":      asset.Name,
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
		asset, err := decodeAsset(kv.Value)
		if err != nil {
			return err
		}
		ids[asset.ID] = true
		if dir.Exists(asset.ID) {
			continue
		}
		err = d.downloadAsset(ctx, asset)
		if err == nil {
			log.Info("asset: downloaded a local copy", map[string]interface{}{
				"name": asset.Name,
				"id":   asset.ID,
			})
			continue
		}

		// continue even when download failed because, if sabakan died,
		// operators could not workaround by, for example, re-uploading
		// the asset.
		log.Error("asset: download failed", map[string]interface{}{
			log.FnError: err,
			"name":      asset.Name,
			"id":        asset.ID,
		})
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

func (d *driver) handleAssetAdd(ctx context.Context, asset *sabakan.Asset) error {
	if d.getAssetDir().Exists(asset.ID) {
		return nil
	}

	err := d.downloadAsset(ctx, asset)
	if err == nil {
		log.Info("asset: downloaded a local copy", map[string]interface{}{
			"name": asset.Name,
			"id":   asset.ID,
		})
	} else {
		// continue even when download failed because, if sabakan died,
		// operators could not workaround by, for example, re-uploading
		// the asset.
		log.Error("asset: download failed", map[string]interface{}{
			log.FnError: err,
			"name":      asset.Name,
			"id":        asset.ID,
		})
	}

	return nil
}

func (d *driver) handleAssetUpdate(ctx context.Context, oldA, newA *sabakan.Asset) error {
	if oldA.ID != newA.ID {
		err := d.handleAssetDelete(ctx, oldA)
		if err != nil {
			return err
		}
	}

	return d.handleAssetAdd(ctx, newA)
}

func (d *driver) handleAssetDelete(ctx context.Context, asset *sabakan.Asset) error {
	dir := d.getAssetDir()

	if !dir.Exists(asset.ID) {
		return nil
	}

	log.Info("asset: delete a local copy", map[string]interface{}{
		"name": asset.Name,
		"id":   asset.ID,
	})
	err := dir.Remove(asset.ID)
	if err != nil {
		log.Error("asset: failed to remove a local copy", map[string]interface{}{
			log.FnError: err,
			"name":      asset.Name,
			"id":        asset.ID,
		})
	}
	return nil
}

func (d *driver) handleAssetEvent(ctx context.Context, ev *clientv3.Event) error {
	switch {
	case ev.IsCreate():
		a, err := decodeAsset(ev.Kv.Value)
		if err != nil {
			return err
		}
		return d.handleAssetAdd(ctx, a)
	case ev.IsModify():
		prevA, err := decodeAsset(ev.PrevKv.Value)
		if err != nil {
			return err
		}
		newA, err := decodeAsset(ev.Kv.Value)
		if err != nil {
			return err
		}
		return d.handleAssetUpdate(ctx, prevA, newA)
	default: // DELETE
		a, err := decodeAsset(ev.PrevKv.Value)
		if err != nil {
			return err
		}
		return d.handleAssetDelete(ctx, a)
	}
}

func (d *driver) startAssetUpdater(ctx context.Context, ch <-chan EventPool) error {
	for ep := range ch {
		jitter := rand.Intn(maxJitterSeconds * 100)
		log.Info("asset updater: waiting...", map[string]interface{}{
			"centiseconds": jitter,
		})
		select {
		case <-time.After(time.Duration(jitter) * 10 * time.Millisecond):
		case <-ctx.Done():
			return nil
		}

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
