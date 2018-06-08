package etcd

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan"
)

type updateData struct {
	index   sabakan.ImageIndex
	deleted []string
}

const (
	maxJitterSeconds = 60
	maxImageURLs     = 10
)

func pullImage(ctx context.Context, client *cmd.HTTPClient, u string) (*http.Response, error) {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	return client.Do(req)
}

func (d *driver) addImageURL(ctx context.Context, os, id string) error {
RETRY:
	index, rev, err := d.imageGetIndexWithRev(ctx, os)
	if err != nil {
		return err
	}

	img := index.Find(id)
	if img == nil {
		// deleted from the index
		return nil
	}

	img.URLs = append(img.URLs, d.myURL("/api/v1/images", os, id))

	indexKey := path.Join(KeyImages, os)
	indexJSON, err := json.Marshal(index)
	if err != nil {
		return err
	}

	resp, err := d.client.Txn(ctx).
		If(
			clientv3.Compare(clientv3.ModRevision(indexKey), "=", rev),
		).
		Then(
			clientv3.OpPut(indexKey, string(indexJSON)),
		).
		Commit()
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		goto RETRY
	}

	return nil
}

func (d *driver) updateImageForOS(ctx context.Context, client *cmd.HTTPClient, os string, data updateData) error {
	dir := d.getImageDir(os)

	err := dir.GC(data.deleted)
	if err != nil {
		return err
	}

OUTER:
	for _, img := range data.index {
		if dir.Exists(img.ID) {
			continue
		}

		urls := make([]string, len(img.URLs))
		for i, v := range rand.Perm(len(urls)) {
			urls[v] = img.URLs[i]
		}

		for _, u := range urls {
			resp, err := pullImage(ctx, client, u)
			if err != nil {
				continue
			}

			err = dir.Extract(resp.Body, img.ID, imageMembers[os])
			resp.Body.Close()
			if err != nil {
				// this is critical
				return err
			}

			log.Info("image updater: pulled image", map[string]interface{}{
				"os":  os,
				"id":  img.ID,
				"url": u,
			})

			if len(img.URLs) < maxImageURLs {
				err = d.addImageURL(ctx, os, img.ID)
				if err != nil {
					return err
				}
			}
			continue OUTER
		}

		log.Error("failed to pull image", map[string]interface{}{
			"os":   os,
			"id":   img.ID,
			"urls": img.URLs,
		})
	}

	return nil
}

func (d *driver) updateImage(ctx context.Context, client *cmd.HTTPClient) error {
	key := KeyImages + "/"
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
		err = d.updateImageForOS(ctx, client, os, data)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *driver) startImageUpdater(ctx context.Context, ch <-chan struct{}) error {
	client := &cmd.HTTPClient{
		Client: &http.Client{},
	}

	for {
		err := d.updateImage(ctx, client)
		if err != nil {
			return err
		}

		select {
		case <-ch:
			jitter := rand.Intn(maxJitterSeconds)
			log.Info("image updater: waiting...", map[string]interface{}{
				"seconds": jitter,
			})
			select {
			case <-time.After(time.Duration(jitter) * time.Second):
			case <-ctx.Done():
				return nil
			}
		case <-ctx.Done():
			return nil
		}
	}
}
