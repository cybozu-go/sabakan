package etcd

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
	"go.etcd.io/etcd/clientv3"
)

func (d *driver) loadLastRev() int64 {
	p := filepath.Join(d.dataDir, LastRevFile)
	f, err := os.Open(p)
	if err != nil {
		log.Warn("failed to open lastrev file", map[string]interface{}{
			log.FnError: err,
		})
		os.Remove(p)
		return 0
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		log.Warn("failed to read lastrev file", map[string]interface{}{
			log.FnError: err,
		})
		os.Remove(p)
		return 0
	}

	rev, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		log.Warn("invalid lastrev file", map[string]interface{}{
			log.FnError: err,
		})
		os.Remove(p)
		return 0
	}

	return rev
}

func (d *driver) saveLastRev(rev int64) error {
	p := filepath.Join(d.dataDir, LastRevFile)
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_SYNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(strconv.FormatInt(rev, 10))
	return err
}

func (d *driver) initStateful(ctx context.Context) (int64, error) {
	// obtain the current revision.
	resp, err := d.client.Get(ctx, "/")
	if err != nil {
		return 0, err
	}
	rev := resp.Header.Revision

	err = d.initAssets(ctx, rev)
	if err != nil {
		return 0, err
	}

	return rev, nil
}

// Helper function to pool events
func poolEvents(ctx context.Context, sendCh chan<- EventPool, recvCh <-chan EventPool) error {
	var ep EventPool

	for {
		if len(ep.Events) > 0 {
			select {
			case sendCh <- ep:
				ep.Events = nil
			case <-ctx.Done():
				return nil
			case ep2 := <-recvCh:
				ep.Rev = ep2.Rev
				ep.Events = append(ep.Events, ep2.Events...)
			}
			continue
		}

		select {
		case <-ctx.Done():
			return nil
		case ep = <-recvCh:
		}
	}
}

// startStatefulWatcher is a goroutine to begin watching for etcd events.
//
// This goroutine keeps the last seen revision in the data directory
// to resume watching between restarts.  All events will be dispatched
// to ch.
func (d *driver) startStatefulWatcher(ctx context.Context, ch chan<- EventPool) error {
	defer close(ch)

RETRY:
	rev := d.loadLastRev()
	if rev == 0 {
		log.Info("initialize stateful watching", nil)

		var err error
		rev, err = d.initStateful(ctx)
		if err != nil {
			return err
		}
	} else {
		log.Info("resume stateful watching", map[string]interface{}{
			"rev": rev,
		})
	}

	rch := d.client.Watch(ctx, KeyAssets,
		clientv3.WithPrefix(),
		clientv3.WithPrevKV(),
		clientv3.WithRev(rev+1),
		clientv3.WithCreatedNotify(),
	)

	// a helper goroutine to pool events
	poolCh := make(chan EventPool)
	env := well.NewEnvironment(ctx)
	env.Go(func(ctx context.Context) error {
		return poolEvents(ctx, ch, poolCh)
	})
	env.Stop()

	for resp := range rch {
		if resp.CompactRevision != 0 {
			log.Warn("database has been compacted; re-initializing", map[string]interface{}{
				"compactedrev": resp.CompactRevision,
			})
			err := d.saveLastRev(0)
			if err != nil {
				return err
			}

			// the watch will be canceled by the server as described in:
			// https://godoc.org/go.etcd.io/etcd/clientv3#Watcher
			for range rch {
			}
			goto RETRY
		}

		if len(resp.Events) == 0 {
			continue
		}

		poolCh <- EventPool{
			Rev:    resp.Header.Revision,
			Events: resp.Events,
		}
	}

	env.Cancel(nil)
	return env.Wait()
}
