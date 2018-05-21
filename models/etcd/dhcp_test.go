package etcd

import (
	"context"
	"encoding/json"
	"net"
	"reflect"
	"testing"

	"github.com/coreos/etcd/clientv3"
	"github.com/cybozu-go/sabakan"
)

var testDHCPConfig = sabakan.DHCPConfig{
	GatewayOffset: 100,
	LeaseMinutes:  30,
}

func testDHCPPutConfig(t *testing.T) {
	d := testNewDriver(t)
	config := &testDHCPConfig
	err := d.putDHCPConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := d.client.Get(context.Background(), t.Name()+KeyDHCP)
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
	d := testNewDriver(t)

	config := &testDHCPConfig

	bytes, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.client.Put(context.Background(), t.Name()+KeyDHCP, string(bytes))
	if err != nil {
		t.Fatal(err)
	}

	actual, err := d.getDHCPConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(config, actual) {
		t.Errorf("unexpected loaded config %#v", actual)
	}
}

func testSetupConfig(t *testing.T, d *driver) {
	ipam := &testIPAMConfig
	config := &testDHCPConfig

	err := d.putIPAMConfig(context.Background(), ipam)
	if err != nil {
		t.Fatal(err)
	}

	err = d.putDHCPConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
}

func testDHCPLease(t *testing.T) {
	d := testNewDriver(t)

	testSetupConfig(t, d)

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

	resp, err := d.client.Get(context.Background(), t.Name()+KeyLeaseUsages, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Kvs) == 0 {
		t.Error("leaseUsage wasn't stored")
	}
}

func testDHCPRenew(t *testing.T) {
	d := testNewDriver(t)

	testSetupConfig(t, d)

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
	d := testNewDriver(t)

	testSetupConfig(t, d)

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

func TestDHCP(t *testing.T) {
	t.Run("Put", testDHCPPutConfig)
	t.Run("Get", testDHCPGetConfig)
	t.Run("Lease", testDHCPLease)
	t.Run("Renew", testDHCPRenew)
	t.Run("Release", testDHCPRelease)
}
