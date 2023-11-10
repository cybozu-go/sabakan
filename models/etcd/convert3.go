package etcd

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan/v3"
	ign22 "github.com/flatcar/ignition/config/v2_2/types"
	"github.com/vincent-petithory/dataurl"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

func (d *driver) convertTo3(ctx context.Context, mu *concurrency.Mutex) error {
	resp, err := d.client.Get(ctx, "ignitions/templates/", clientv3.WithPrefix())
	if err != nil {
		return err
	}

	var ops []clientv3.Op
	tmpls := make(map[string]*sabakan.IgnitionTemplate)

	for _, kv := range resp.Kvs {
		var ign ign22.Config
		key := string(kv.Key[len("ignitions/templates/"):])
		err := json.Unmarshal(kv.Value, &ign)
		if err != nil {
			return errors.New("invalid template: " + key)
		}
		for i := range ign.Storage.Files {
			file := &ign.Storage.Files[i]
			file.Contents.Source = "data:," + dataurl.EscapeString(file.Contents.Source)
		}
		data, err := json.Marshal(ign)
		if err != nil {
			return err
		}
		tmpls[key] = &sabakan.IgnitionTemplate{
			Version:  sabakan.Ignition2_2,
			Template: json.RawMessage(data),
		}
		ops = append(ops, clientv3.OpDelete(string(kv.Key)))
	}

	resp, err = d.client.Get(ctx, "ignitions/meta/", clientv3.WithPrefix())
	if err != nil {
		return err
	}
	for _, kv := range resp.Kvs {
		var metadata map[string]interface{}
		key := string(kv.Key[len("ignitions/meta/"):])
		err := json.Unmarshal(kv.Value, &metadata)
		if err != nil {
			return err
		}
		ops = append(ops, clientv3.OpDelete(string(kv.Key)))
		tmpl, ok := tmpls[key]
		if !ok {
			log.Warn("dangling meta data for "+key, nil)
			continue
		}
		tmpl.Metadata = metadata
	}

	for key, tmpl := range tmpls {
		data, err := json.Marshal(tmpl)
		if err != nil {
			return err
		}
		ops = append(ops, clientv3.OpPut(KeyIgnitions+key, string(data)))
	}

	// update schema version
	const thisVersion = "3"
	ops = append(ops, clientv3.OpPut(KeyVersion, thisVersion))

	tresp, err := d.client.Txn(ctx).
		If(mu.IsOwner()).
		Then(ops...).
		Commit()
	if err != nil {
		return err
	}
	if !tresp.Succeeded {
		return errLostOwner
	}

	log.Info("updated schema version", map[string]interface{}{
		"to": thisVersion,
	})
	return nil
}
