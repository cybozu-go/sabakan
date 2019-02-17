package etcd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan/v2"
)

func auditKey(t time.Time) string {
	return KeyAudit + t.UTC().Format("20060102") + "/"
}

func (d *driver) addLog(ctx context.Context, ts time.Time, rev int64, cat sabakan.AuditCategory,
	instance, action, detail string) {

	a := sabakan.NewAuditLog(ctx, ts, rev, cat, instance, action, detail)
	j, err := json.Marshal(a)
	if err != nil { // unlikely
		log.Error("etcd: addLog failed", map[string]interface{}{
			log.FnError: err,
			"category":  string(cat),
			"instance":  instance,
			"action":    action,
		})
		return
	}

	key := auditKey(ts) + fmt.Sprintf("%016x", uint64(rev))
	_, err = d.client.Put(ctx, key, string(j))
	if err == nil {
		return
	}

	log.Error("etcd: addLog failed", map[string]interface{}{
		log.FnError: err,
		"category":  string(cat),
		"instance":  instance,
		"action":    action,
	})
}

func (d *driver) logLastGCTime(ctx context.Context, nowData string) (t time.Time, rev int64, e error) {
RETRY:
	resp, err := d.client.Get(ctx, KeyAuditLastGC)
	if err != nil {
		e = err
		return
	}
	if resp.Count == 0 {
		_, err := d.client.Txn(ctx).
			If(clientv3util.KeyMissing(KeyAuditLastGC)).
			Then(clientv3.OpPut(KeyAuditLastGC, nowData)).
			Commit()
		if err != nil {
			e = err
			return
		}
		goto RETRY
	}

	err = json.Unmarshal(resp.Kvs[0].Value, &t)
	if err != nil {
		e = err
		return
	}

	rev = resp.Kvs[0].ModRevision
	return
}

func (d *driver) logCompact(ctx context.Context, now time.Time) error {
	oldest := now.Add(time.Duration(-logRetentionDays) * 24 * time.Hour)
	key := auditKey(oldest)

	log.Info("log: compacting...", map[string]interface{}{
		"key": key,
	})

	resp, err := d.client.Delete(ctx, KeyAudit, clientv3.WithRange(key))
	if err != nil {
		return err
	}

	log.Info("log: compacted", map[string]interface{}{
		"deleted": resp.Deleted,
	})

	return nil
}

func (d *driver) logTryCompact(ctx context.Context, now time.Time) error {
	j, err := json.Marshal(now)
	if err != nil {
		return err
	}
	nowData := string(j)

	lastGC, rev, err := d.logLastGCTime(ctx, nowData)
	if err != nil {
		return err
	}

	if now.Sub(lastGC) < logCompactionInterval {
		return nil
	}

	tresp, err := d.client.Txn(ctx).
		If(clientv3.Compare(clientv3.ModRevision(KeyAuditLastGC), "=", rev)).
		Then(clientv3.OpPut(KeyAuditLastGC, nowData)).
		Commit()
	if err != nil {
		return err
	}
	if !tresp.Succeeded {
		return nil
	}

	return d.logCompact(ctx, now)
}

// logCompactor is a goroutine to compact logs periodically.
func (d *driver) logCompactor(ctx context.Context) error {
	ticker := time.NewTicker(logCompactionTick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case now := <-ticker.C:
			err := d.logTryCompact(ctx, now.UTC())
			if err != nil {
				return err
			}
		}
	}
}

func (d *driver) logDump(ctx context.Context, since, until time.Time, w io.Writer) error {
	bufw := bufio.NewWriterSize(w, 2048)

	key := KeyAudit
	endKey := clientv3.GetPrefixRangeEnd(KeyAudit)

	if !since.IsZero() {
		key = auditKey(since)
	}
	if !until.IsZero() {
		endKey = auditKey(until)
	}

	// paginate for large number of logs

	resp, err := d.client.Get(ctx, key,
		clientv3.WithRange(endKey),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
		clientv3.WithLimit(logPageSize),
	)
	if err != nil {
		return err
	}
	rev := resp.Header.Revision //  to retrieve following pages at the same revision.
	kvs := resp.Kvs

REDO:
	for _, kv := range kvs {
		_, err = bufw.Write(kv.Value)
		if err != nil {
			return err
		}
		err = bufw.WriteByte('\n')
		if err != nil {
			return err
		}
	}

	if resp.More {
		resp, err = d.client.Get(ctx, string(resp.Kvs[len(resp.Kvs)-1].Key),
			clientv3.WithRange(endKey),
			clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
			clientv3.WithLimit(logPageSize),
			clientv3.WithRev(rev),
		)
		if err != nil {
			return err
		}

		// ignore the first key
		kvs = resp.Kvs[1:]
		goto REDO
	}

	return bufw.Flush()
}

type logDriver struct {
	*driver
}

func (d logDriver) Dump(ctx context.Context, since, until time.Time, w io.Writer) error {
	return d.logDump(ctx, since, until, w)
}
