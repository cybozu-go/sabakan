package mock

import (
	"context"
	"errors"
	"net"

	"github.com/cybozu-go/sabakan"
)

type leaseUsage struct {
	leaseRange *sabakan.LeaseRange
	macMap     map[string]int // MAC address to index-in-range
	usageMap   map[int]bool
}

func (l *leaseUsage) lease(mac net.HardwareAddr) (net.IP, error) {
	if idx, ok := l.macMap[mac.String()]; ok {
		return l.leaseRange.IP(idx), nil
	}

	for i := 0; i < l.leaseRange.Count; i++ {
		if l.usageMap[i] {
			continue
		}
		l.usageMap[i] = true
		l.macMap[mac.String()] = i
		return l.leaseRange.IP(i), nil
	}

	return nil, errors.New("no leasable IP address found from " + l.leaseRange.Key())
}

func (l *leaseUsage) renew(mac net.HardwareAddr) error {
	_, ok := l.macMap[mac.String()]
	if !ok {
		return errors.New("not leased for " + mac.String())
	}
	return nil
}

func (l *leaseUsage) release(mac net.HardwareAddr) {
	key := mac.String()

	idx, ok := l.macMap[key]
	if !ok {
		return
	}

	delete(l.macMap, key)
	delete(l.usageMap, idx)
}

func newLeaseUsage(lr *sabakan.LeaseRange) *leaseUsage {
	return &leaseUsage{
		leaseRange: lr,
		macMap:     make(map[string]int),
		usageMap:   make(map[int]bool),
	}
}

func (d *driver) putDHCPConfig(ctx context.Context, config *sabakan.DHCPConfig) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	copied := *config
	d.dhcp = &copied
	return nil
}

func (d *driver) getDHCPConfig() (*sabakan.DHCPConfig, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.dhcp == nil {
		return nil, errors.New("DHCPConfig is not set")
	}
	copied := *d.dhcp
	return &copied, nil
}

func (d *driver) dhcpLease(ctx context.Context, ifaddr net.IP, mac net.HardwareAddr) (net.IP, error) {
	ipam, err := d.getIPAMConfig()
	if err != nil {
		return nil, err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	lr := ipam.LeaseRange(ifaddr)
	if lr == nil {
		return nil, errors.New("invalid ifaddr: " + ifaddr.String())
	}

	key := lr.Key()
	lu := d.leases[key]
	if lu == nil {
		lu = newLeaseUsage(lr)
		d.leases[key] = lu
	}

	return lu.lease(mac)
}

func (d *driver) dhcpRenew(ctx context.Context, ciaddr net.IP, mac net.HardwareAddr) error {
	ipam, err := d.getIPAMConfig()
	if err != nil {
		return err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	lr := ipam.LeaseRange(ciaddr)
	if lr == nil {
		return errors.New("invalid ciaddr: " + ciaddr.String())
	}

	key := lr.Key()
	lu := d.leases[key]
	if lu == nil {
		return errors.New("not leased for " + mac.String())
	}
	return lu.renew(mac)
}

func (d *driver) dhcpRelease(ctx context.Context, ciaddr net.IP, mac net.HardwareAddr) error {
	ipam, err := d.getIPAMConfig()
	if err != nil {
		return err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	lr := ipam.LeaseRange(ciaddr)
	if lr == nil {
		return errors.New("invalid ciaddr: " + ciaddr.String())
	}

	key := lr.Key()
	lu := d.leases[key]
	if lu != nil {
		lu.release(mac)
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
