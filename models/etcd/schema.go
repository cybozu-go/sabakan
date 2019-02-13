package etcd

import (
	"context"
	"errors"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan"
)

const noVersion = "1"

func (d *driver) Version(ctx context.Context) (string, error) {
	resp, err := d.client.Get(ctx, KeyVersion)
	if err != nil {
		return "", err
	}

	if resp.Count == 0 {
		return noVersion, nil
	}

	return string(resp.Kvs[0].Value), nil
}

func (d *driver) Upgrade(ctx context.Context) error {
	sess, err := concurrency.NewSession(d.client)
	if err != nil {
		return err
	}
	defer sess.Close()

	mu := concurrency.NewMutex(sess, KeySchemaLockPrefix)
	if err := mu.Lock(ctx); err != nil {
		return err
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		mu.Unlock(ctx)
		cancel()
	}()

	sv, err := d.Version(ctx)
	if err != nil {
		return err
	}

	if sv == sabakan.SchemaVersion {
		return nil
	}

	log.Info("upgrading schema version", map[string]interface{}{
		"from": sv,
		"to":   sabakan.SchemaVersion,
	})

	switch sv {
	case "1":
		err := d.convertTo2(ctx, mu)
		if err != nil {
			return err
		}

		// fallthrough when case "2" is added
		//fallthrough
	default:
		return errors.New("unknown schema version: " + sv)
	}

	return nil
}
