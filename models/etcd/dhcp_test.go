package etcd

import (
	"context"
	"encoding/json"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/cybozu-go/sabakan/v2"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var testDHCPConfig = sabakan.DHCPConfig{
	LeaseMinutes: 30,
	DNSServers:   []string{"10.0.0.1", "10.0.0.2"},
}

func testDHCPPutConfig(t *testing.T) {
	d, ch := testNewDriver(t)
	config := &testDHCPConfig
	err := d.putDHCPConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	<-ch

	resp, err := d.client.Get(context.Background(), KeyDHCP)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Kvs) != 1 {
		t.Error("config was not saved")
	}
	var actual sabakan.DHCPConfig
	err = json.Unmarshal(resp.Kvs[0].Value, &actual)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(config, &actual) {
		t.Errorf("unexpected saved config %#v", actual)
	}
}

func testDHCPGetConfig(t *testing.T) {
	d, ch := testNewDriver(t)
	config := &testDHCPConfig

	bytes, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.client.Put(context.Background(), KeyDHCP, string(bytes))
	if err != nil {
		t.Fatal(err)
	}
	<-ch

	actual, err := d.getDHCPConfig()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(config, actual) {
		t.Errorf("unexpected loaded config %#v", actual)
	}
}

func testSetupConfig(t *testing.T, d *driver, ch <-chan struct{}) {
	ipam := &testIPAMConfig
	config := &testDHCPConfig

	err := d.putIPAMConfig(context.Background(), ipam)
	if err != nil {
		t.Fatal(err)
	}
	<-ch

	err = d.putDHCPConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	<-ch
}

func testDHCPLease(t *testing.T) {
	d, ch := testNewDriver(t)
	testSetupConfig(t, d, ch)

	interfaceip := net.ParseIP("10.69.0.195")
	mac := net.HardwareAddr([]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66})

	dhcpip, err := d.dhcpLease(context.Background(), interfaceip, mac)
	if err != nil {
		t.Fatal(err)
	}

	expected := net.ParseIP("10.69.0.224")
	if !dhcpip.Equal(expected) {
		t.Error("dhcpip is not expected: ", dhcpip.String())
	}

	mac2 := net.HardwareAddr([]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x67})
	dhcpip, err = d.dhcpLease(context.Background(), interfaceip, mac2)
	if err != nil {
		t.Fatal(err)
	}
	expected2 := net.ParseIP("10.69.0.225")
	if !dhcpip.Equal(expected2) {
		t.Error("dhcpip is not expected: ", dhcpip.String())
	}

	dhcpip, err = d.dhcpLease(context.Background(), interfaceip, mac)
	if err != nil {
		t.Fatal(err)
	}
	if !dhcpip.Equal(expected) {
		t.Error("dhcpip is not expected: ", dhcpip.String())
	}

	for i := 0; i < 29; i++ {
		mac := net.HardwareAddr([]byte{0x11, 0x22, 0x33, 0x44, 0x55, byte(i)})
		_, err := d.dhcpLease(context.Background(), interfaceip, mac)
		if err != nil {
			t.Fatal(err)
		}
	}

	mac3 := net.HardwareAddr([]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0xFF})
	_, err = d.dhcpLease(context.Background(), interfaceip, mac3)
	if err == nil {
		t.Error("dhcp lease should fail")
	}

	resp, err := d.client.Get(context.Background(), KeyLeaseUsages, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Kvs) == 0 {
		t.Error("leaseUsage wasn't stored")
	}
}

func testDHCPRenew(t *testing.T) {
	d, ch := testNewDriver(t)
	testSetupConfig(t, d, ch)

	leasedip := net.ParseIP("10.69.0.224")
	mac := net.HardwareAddr([]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66})
	err := d.dhcpRenew(context.Background(), leasedip, mac)
	if err == nil {
		t.Error("dhcpRenew should fail")
	}

	interfaceip := net.ParseIP("10.69.0.195")
	_, err = d.dhcpLease(context.Background(), interfaceip, mac)
	if err != nil {
		t.Fatal(err)
	}

	err = d.dhcpRenew(context.Background(), leasedip, mac)
	if err != nil {
		t.Error(err)
	}
}

func testDHCPRelease(t *testing.T) {
	d, ch := testNewDriver(t)
	testSetupConfig(t, d, ch)

	interfaceip := net.ParseIP("10.69.0.195")
	mac := net.HardwareAddr([]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66})

	dhcpip, err := d.dhcpLease(context.Background(), interfaceip, mac)
	if err != nil {
		t.Fatal(err)
	}

	err = d.dhcpRelease(context.Background(), dhcpip, mac)
	if err != nil {
		t.Error(err)
	}

	mac2 := net.HardwareAddr([]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x67})
	dhcpip2, err := d.dhcpLease(context.Background(), interfaceip, mac2)
	if err != nil {
		t.Fatal(err)
	}

	if !dhcpip.Equal(dhcpip2) {
		t.Error("unexpected ip address", dhcpip, dhcpip2)
	}
}

func testDHCPDecline(t *testing.T) {
	d, ch := testNewDriver(t)
	testSetupConfig(t, d, ch)

	interfaceip := net.ParseIP("10.69.0.195")
	mac := net.HardwareAddr([]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66})

	dhcpip, err := d.dhcpLease(context.Background(), interfaceip, mac)
	if err != nil {
		t.Fatal(err)
	}

	err = d.dhcpDecline(context.Background(), dhcpip, mac)
	if err != nil {
		t.Error(err)
	}

	dhcpip2, err := d.dhcpLease(context.Background(), interfaceip, mac)
	if err != nil {
		t.Fatal(err)
	}

	if dhcpip.Equal(dhcpip2) {
		t.Error("declined IP address should not be used until expired")
	}
}

func testDHCPLeaseExpiration(t *testing.T) {
	d, ch := testNewDriver(t)
	testSetupConfig(t, d, ch)

	ipam, err := d.getIPAMConfig()
	if err != nil {
		t.Fatal(err)
	}

	interfaceip := net.ParseIP("10.69.0.195")

	// prepare data
	mac := net.HardwareAddr([]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66})
	dhcpip, err := d.dhcpLease(context.Background(), interfaceip, mac)
	if err != nil {
		t.Fatal(err)
	}

	lr := ipam.LeaseRange(interfaceip)
	lrkey := lr.Key()

RETRY:
	// retrieve lease information and update expiration to 2000-01-01
	lu, err := d.getLeaseUsage(context.Background(), lrkey)
	if err != nil {
		t.Fatal(err)
	}

	li := lu.hwMap[mac.String()]
	li.LeaseUntil = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	lu.hwMap[mac.String()] = li

	succeeded, err := d.updateLeaseUsage(context.Background(), lrkey, lu)
	if err != nil {
		t.Fatal(err)
	}
	if !succeeded {
		goto RETRY
	}

	// register machine to reuse expired address
	mac2 := net.HardwareAddr([]byte{0x22, 0x33, 0x44, 0x55, 0x66, 0x77})
	dhcpip2, err := d.dhcpLease(context.Background(), interfaceip, mac2)
	if err != nil {
		t.Fatal(err)
	}
	if !dhcpip2.Equal(dhcpip) {
		t.Error("expired address was not reused")
	}
}

func testDHCPLeaseRace(t *testing.T) {
	d, ch := testNewDriver(t)
	testSetupConfig(t, d, ch)
	ipam, err := d.getIPAMConfig()
	if err != nil {
		t.Fatal(err)
	}

	interfaceip := net.ParseIP("10.69.0.195")
	lr := ipam.LeaseRange(interfaceip)
	lrkey := lr.Key()

RETRY:
	// retrieve usage data #1 and #2 with revision, and update data on memory
	lu1, err := d.getLeaseUsage(context.Background(), lrkey)
	if err != nil {
		t.Fatal(err)
	}

	lu2, err := d.getLeaseUsage(context.Background(), lrkey)
	if err != nil {
		t.Fatal(err)
	}

	// update data#2 on etcd; this increments revision
	succeeded2, err := d.updateLeaseUsage(context.Background(), lrkey, lu2)
	if err != nil {
		t.Fatal(err)
	}
	if !succeeded2 {
		goto RETRY
	}

	// try to update data#1 on etcd; this must fail
	succeeded1, err := d.updateLeaseUsage(context.Background(), lrkey, lu1)
	if err != nil {
		t.Fatal(err)
	}
	if succeeded1 {
		t.Error("update operations should fail, if revision number has been changed")
	}
}

func testDummyMAC(t *testing.T) {
	t.Parallel()

	mac := generateDummyMAC(257)
	if mac.String() != "ff:00:00:00:01:01" {
		t.Error("unexpected MAC address", mac)
	}
}

func TestDHCP(t *testing.T) {
	t.Run("Put", testDHCPPutConfig)
	t.Run("Get", testDHCPGetConfig)
	t.Run("Lease", testDHCPLease)
	t.Run("Renew", testDHCPRenew)
	t.Run("Release", testDHCPRelease)
	t.Run("Decline", testDHCPDecline)
	t.Run("Expire", testDHCPLeaseExpiration)
	t.Run("Race", testDHCPLeaseRace)

	t.Run("Generate Dummy MAC", testDummyMAC)
}
