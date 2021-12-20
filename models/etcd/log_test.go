package etcd

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/cybozu-go/sabakan/v2"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func testLogAdd(t *testing.T) {
	t.Parallel()

	d, _ := testNewDriver(t)

	ctx := context.Background()
	ts := time.Date(2013, time.April, 5, 1, 2, 3, 4, time.UTC)
	d.addLog(ctx, ts, 255, sabakan.AuditIPAM, "config", "put", "test")

	key := "audit/20130405/00000000000000ff"
	resp, err := d.client.Get(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Count == 0 {
		t.Fatal("log entry not found at", key)
	}

	var a sabakan.AuditLog
	err = json.Unmarshal(resp.Kvs[0].Value, &a)
	if err != nil {
		t.Fatal(err)
	}
	if !a.Timestamp.Equal(ts) {
		t.Error(a.Timestamp.String(), "!=", ts.String())
	}
}

func testLogCompact(t *testing.T) {
	t.Parallel()

	d, _ := testNewDriver(t)
	ctx := context.Background()

	now := time.Date(2013, time.April, 5, 1, 2, 3, 4, time.UTC)
	ts := now.Add(-61 * 24 * time.Hour)
	for i := int64(0); i < 3; i++ {
		d.addLog(ctx, ts, 100+i, sabakan.AuditIPAM, "config", "put", "test")
		ts = ts.Add(24 * time.Hour)
	}

	countFunc := func() (int64, error) {
		resp, err := d.client.Get(ctx, KeyAudit, clientv3.WithPrefix(), clientv3.WithCountOnly())
		if err != nil {
			return 0, err
		}
		return resp.Count, nil
	}

	count, err := countFunc()
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Error(`count != 3`, count)
	}

	err = d.logCompact(ctx, now)
	if err != nil {
		t.Fatal(err)
	}
	count, err = countFunc()
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Error(`count != 2`, count)
	}

	err = d.logCompact(ctx, now.Add(48*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	count, err = countFunc()
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Error(`count != 0`, count)
	}
}

func testLogTryCompact(t *testing.T) {
	t.Parallel()

	d, _ := testNewDriver(t)
	ctx := context.Background()

	now := time.Date(2013, time.April, 5, 1, 2, 3, 4, time.UTC)
	ts := now.Add(-61 * 24 * time.Hour)
	for i := int64(0); i < 3; i++ {
		d.addLog(ctx, ts, 100+i, sabakan.AuditIPAM, "config", "put", "test")
		ts = ts.Add(24 * time.Hour)
	}

	countFunc := func() (int64, error) {
		resp, err := d.client.Get(ctx, KeyAudit, clientv3.WithPrefix(), clientv3.WithCountOnly())
		if err != nil {
			return 0, err
		}
		return resp.Count, nil
	}

	err := d.logTryCompact(ctx, now)
	if err != nil {
		t.Fatal(err)
	}
	count, err := countFunc()
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Error(`count != 3`, count)
	}

	resp, err := d.client.Get(ctx, KeyAuditLastGC)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Count == 0 {
		t.Fatal(`resp.Count == 0`)
	}
	var lastGC time.Time
	err = json.Unmarshal(resp.Kvs[0].Value, &lastGC)
	if err != nil {
		t.Fatal(err)
	}
	if !lastGC.Equal(now) {
		t.Error(`!lastGC.Equal(now)`, lastGC)
	}

	now = now.Add(24 * time.Hour)
	err = d.logTryCompact(ctx, now)
	if err != nil {
		t.Fatal(err)
	}
	count, err = countFunc()
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Error(`count != 1`, count)
	}
}

func testLogDump(t *testing.T) {
	t.Parallel()

	d, _ := testNewDriver(t)
	ctx := context.Background()

	begin := time.Date(2013, time.April, 5, 1, 2, 3, 4, time.UTC)
	end := begin
	for i := int64(0); i < 2*logPageSize; i++ {
		d.addLog(ctx, end, 100+i, sabakan.AuditIPAM, "config", "put", "test")
		end = end.Add(time.Hour)
	}

	buf := new(bytes.Buffer)
	err := d.logDump(ctx, time.Time{}, time.Time{}, buf)
	if err != nil {
		t.Fatal(err)
	}
	data := buf.Bytes()
	count := bytes.Count(data, []byte("\n"))
	if count != 2*logPageSize {
		t.Error(`count != 2 * logPageSize`, count)
	}

	var a sabakan.AuditLog
	err = json.Unmarshal(data[:bytes.IndexByte(data, '\n')], &a)
	if err != nil {
		t.Fatal(err)
	}
	if a.Revision != 100 {
		t.Error(`a.Revision != 100`, a.Revision)
	}
	err = json.Unmarshal(data[bytes.LastIndexByte(data[:len(data)-1], '\n'):], &a)
	if err != nil {
		t.Fatal(err)
	}
	if a.Revision != 100+2*logPageSize-1 {
		t.Error(`a.Revision != 100+2*logPageSize-1`, a.Revision)
	}

	buf.Reset()
	err = d.logDump(ctx, time.Date(2013, time.April, 6, 0, 0, 0, 0, time.UTC), time.Time{}, buf)
	if err != nil {
		t.Fatal(err)
	}
	data = buf.Bytes()
	count = bytes.Count(data, []byte("\n"))
	if count != 2*logPageSize-23 {
		t.Error(`count != 2*logPageSize-23`, count)
	}

	buf.Reset()
	err = d.logDump(ctx, time.Time{}, time.Date(2013, time.April, 6, 0, 0, 0, 0, time.UTC), buf)
	if err != nil {
		t.Fatal(err)
	}
	data = buf.Bytes()
	count = bytes.Count(data, []byte("\n"))
	if count != 23 {
		t.Error(`count != 23`, count)
	}

	buf.Reset()
	err = d.logDump(ctx,
		time.Date(2013, time.April, 6, 0, 0, 0, 0, time.UTC),
		time.Date(2013, time.April, 7, 0, 0, 0, 0, time.UTC),
		buf)
	if err != nil {
		t.Fatal(err)
	}
	data = buf.Bytes()
	count = bytes.Count(data, []byte("\n"))
	if count != 24 {
		t.Error(`count != 24`, count)
	}
}

func TestLog(t *testing.T) {
	t.Run("Add", testLogAdd)
	t.Run("Compact", testLogCompact)
	t.Run("TryCompact", testLogTryCompact)
	t.Run("Dump", testLogDump)
}
