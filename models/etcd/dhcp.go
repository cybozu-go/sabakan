package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"path"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan"
)

func (d *driver) putDHCPConfig(ctx context.Context, config *sabakan.DHCPConfig) error {
	j, err := json.Marshal(config)
	if err != nil {
		return err
	}

	_, err = d.client.Put(ctx, KeyDHCP, string(j))
	return err
}

func (d *driver) getDHCPConfig() (*sabakan.DHCPConfig, error) {
	v := d.dhcpConfig.Load()
	if v == nil {
		return nil, errors.New("DHCPConfig is not set")
	}

	return v.(*sabakan.DHCPConfig), nil
}

func (d *driver) handleDHCPConfig(ev *clientv3.Event) error {
	if ev.Type == clientv3.EventTypeDelete {
		return nil
	}

	config := new(sabakan.DHCPConfig)
	err := json.Unmarshal(ev.Kv.Value, config)
	if err != nil {
		return err
	}

	d.dhcpConfig.Store(config)
	return nil
}

type leaseInfo struct {
	Index      int       `json:"index"`
	LeaseUntil time.Time `json:"lease"`
}

type leaseUsage struct {
	hwMap    map[string]leaseInfo
	usageMap map[int]bool
	revision int64
}

func (l *leaseUsage) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.hwMap)
}

func (l *leaseUsage) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &l.hwMap)
	if err != nil {
		return err
	}

	um := make(map[int]bool)
	for _, v := range l.hwMap {
		um[v.Index] = true
	}
	l.usageMap = um

	return nil
}

func (l *leaseUsage) gc() {
	now := time.Now()

	for k, v := range l.hwMap {
		if !v.LeaseUntil.Before(now) {
			continue
		}

		// this is safe.
		// ref. https://stackoverflow.com/a/23230406/1493661
		delete(l.usageMap, v.Index)
		delete(l.hwMap, k)
	}
}

func (l *leaseUsage) lease(mac net.HardwareAddr, lr *sabakan.LeaseRange, du time.Duration) (net.IP, error) {
	hwAddr := mac.String()
	leaseUntil := time.Now().Add(du)
	if v, ok := l.hwMap[hwAddr]; ok {
		v.LeaseUntil = leaseUntil
		return lr.IP(v.Index), nil
	}

	l.gc()

	for i := 0; i < lr.Count; i++ {
		if l.usageMap[i] {
			continue
		}
		l.usageMap[i] = true
		l.hwMap[hwAddr] = leaseInfo{i, leaseUntil}
		log.Debug("etcd/dhcp: lease", map[string]interface{}{
			"node_index":  i,
			"mac":         hwAddr,
			"ip":          lr.IP(i),
			"lease_until": leaseUntil,
		})
		return lr.IP(i), nil
	}

	return nil, errors.New("no leasable IP address found from " + lr.Key())
}

func (l *leaseUsage) renew(mac net.HardwareAddr, du time.Duration) error {
	hwAddr := mac.String()
	v, ok := l.hwMap[hwAddr]
	if !ok {
		return errors.New("not leased for " + hwAddr)
	}

	leaseUntil := time.Now().Add(du)
	v.LeaseUntil = leaseUntil
	log.Debug("etcd/dhcp: renew", map[string]interface{}{
		"node_index":  v.Index,
		"mac":         hwAddr,
		"lease_until": leaseUntil,
	})
	return nil
}

func (l *leaseUsage) release(mac net.HardwareAddr) {
	hwAddr := mac.String()

	v, ok := l.hwMap[hwAddr]
	if !ok {
		return
	}

	log.Debug("etcd/dhcp: release", map[string]interface{}{
		"node_index": v.Index,
		"mac":        hwAddr,
	})
	delete(l.usageMap, v.Index)
	delete(l.hwMap, hwAddr)
}

func (l *leaseUsage) decline(mac net.HardwareAddr) {
	hwAddr := mac.String()

	v, ok := l.hwMap[hwAddr]
	if !ok {
		return
	}

	log.Debug("etcd/dhcp: decline", map[string]interface{}{
		"node_index": v.Index,
		"mac":        hwAddr,
	})

	declineKey := generateDummyMAC(v.Index).String()
	l.hwMap[declineKey] = v
	delete(l.hwMap, hwAddr)
}

func generateDummyMAC(idx int) net.HardwareAddr {
	return net.HardwareAddr{
		0xff,
		0,
		byte((idx / 256 / 256 / 256) % 256),
		byte((idx / 256 / 256) % 256),
		byte((idx / 256) % 256),
		byte(idx % 256),
	}
}

func (d *driver) leaseUsageKey(lrkey string) string {
	return path.Join(KeyLeaseUsages, lrkey)
}

func (d *driver) initializeLeaseUsage(ctx context.Context, lrkey string) error {
	var usage leaseUsage
	j, err := json.Marshal(usage)
	if err != nil {
		return err
	}

	key := d.leaseUsageKey(lrkey)
	_, err = d.client.Txn(ctx).
		If(clientv3util.KeyMissing(key)).
		Then(clientv3.OpPut(key, string(j))).
		Else().
		Commit()

	return err
}

func (d *driver) getLeaseUsage(ctx context.Context, lrkey string) (*leaseUsage, error) {
RETRY:
	key := d.leaseUsageKey(lrkey)
	resp, err := d.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		err = d.initializeLeaseUsage(ctx, lrkey)
		if err != nil {
			return nil, err
		}
		goto RETRY
	}

	kv := resp.Kvs[0]
	usage := new(leaseUsage)
	err = json.Unmarshal(kv.Value, usage)
	if err != nil {
		return nil, err
	}

	usage.revision = kv.ModRevision

	return usage, nil
}

func (d *driver) updateLeaseUsage(ctx context.Context, lrkey string, lu *leaseUsage) (bool, error) {
	key := d.leaseUsageKey(lrkey)
	j, err := json.Marshal(lu)
	if err != nil {
		return false, err
	}

	tresp, err := d.client.Txn(ctx).
		If(
			clientv3.Compare(clientv3.ModRevision(key), "=", lu.revision),
		).
		Then(
			clientv3.OpPut(key, string(j)),
		).
		Else().
		Commit()
	if err != nil {
		return false, err
	}

	return tresp.Succeeded, nil
}

func (d *driver) dhcpLease(ctx context.Context, ifaddr net.IP, mac net.HardwareAddr) (net.IP, error) {
	ipam, err := d.getIPAMConfig()
	if err != nil {
		return nil, err
	}

	dc, err := d.getDHCPConfig()
	if err != nil {
		return nil, err
	}

	lr := ipam.LeaseRange(ifaddr)
	if lr == nil {
		return nil, errors.New("invalid ifaddr: " + ifaddr.String())
	}

	lrkey := lr.Key()

RETRY:
	lu, err := d.getLeaseUsage(ctx, lrkey)
	if err != nil {
		return nil, err
	}

	ip, err := lu.lease(mac, lr, dc.LeaseDuration())
	if err != nil {
		return nil, err
	}

	succeeded, err := d.updateLeaseUsage(ctx, lrkey, lu)
	if err != nil {
		return nil, err
	}
	if !succeeded {
		log.Info("etcd: revision mismatch; retrying...", nil)
		goto RETRY
	}

	return ip, nil
}

func (d *driver) dhcpRenew(ctx context.Context, ciaddr net.IP, mac net.HardwareAddr) error {
	ipam, err := d.getIPAMConfig()
	if err != nil {
		return err
	}

	dc, err := d.getDHCPConfig()
	if err != nil {
		return err
	}

	lr := ipam.LeaseRange(ciaddr)
	if lr == nil {
		return errors.New("invalid ciaddr: " + ciaddr.String())
	}

	lrkey := lr.Key()

RETRY:
	lu, err := d.getLeaseUsage(ctx, lrkey)
	if err != nil {
		return err
	}

	err = lu.renew(mac, dc.LeaseDuration())
	if err != nil {
		return err
	}

	succeeded, err := d.updateLeaseUsage(ctx, lrkey, lu)
	if err != nil {
		return err
	}
	if !succeeded {
		log.Info("etcd: revision mismatch; retrying...", nil)
		goto RETRY
	}

	return nil
}

func (d *driver) dhcpRelease(ctx context.Context, ciaddr net.IP, mac net.HardwareAddr) error {
	ipam, err := d.getIPAMConfig()
	if err != nil {
		return err
	}

	lr := ipam.LeaseRange(ciaddr)
	if lr == nil {
		return errors.New("invalid ciaddr: " + ciaddr.String())
	}

	lrkey := lr.Key()

RETRY:
	lu, err := d.getLeaseUsage(ctx, lrkey)
	if err != nil {
		return err
	}

	lu.release(mac)

	succeeded, err := d.updateLeaseUsage(ctx, lrkey, lu)
	if err != nil {
		return err
	}
	if !succeeded {
		log.Info("etcd: revision mismatch; retrying...", nil)
		goto RETRY
	}

	return nil
}

func (d *driver) dhcpDecline(ctx context.Context, ciaddr net.IP, mac net.HardwareAddr) error {
	ipam, err := d.getIPAMConfig()
	if err != nil {
		return err
	}

	lr := ipam.LeaseRange(ciaddr)
	if lr == nil {
		return errors.New("invalid ciaddr: " + ciaddr.String())
	}

	lrkey := lr.Key()

RETRY:
	lu, err := d.getLeaseUsage(ctx, lrkey)
	if err != nil {
		return err
	}

	lu.decline(mac)

	succeeded, err := d.updateLeaseUsage(ctx, lrkey, lu)
	if err != nil {
		return err
	}
	if !succeeded {
		log.Info("etcd: revision mismatch; retrying...", nil)
		goto RETRY
	}

	return nil
}

type dhcpDriver struct {
	*driver
}

func (d dhcpDriver) PutConfig(ctx context.Context, config *sabakan.DHCPConfig) error {
	return d.putDHCPConfig(ctx, config)
}

func (d dhcpDriver) GetConfig() (*sabakan.DHCPConfig, error) {
	return d.getDHCPConfig()
}

func (d dhcpDriver) Lease(ctx context.Context, ifaddr net.IP, mac net.HardwareAddr) (net.IP, error) {
	return d.dhcpLease(ctx, ifaddr, mac)
}

func (d dhcpDriver) Renew(ctx context.Context, ciaddr net.IP, mac net.HardwareAddr) error {
	return d.dhcpRenew(ctx, ciaddr, mac)
}

func (d dhcpDriver) Release(ctx context.Context, ciaddr net.IP, mac net.HardwareAddr) error {
	return d.dhcpRelease(ctx, ciaddr, mac)
}

func (d dhcpDriver) Decline(ctx context.Context, ciaddr net.IP, mac net.HardwareAddr) error {
	return d.dhcpDecline(ctx, ciaddr, mac)
}
