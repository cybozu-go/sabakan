package sabakan

import (
	"net"
	"testing"
	"time"
)

func testGatewayAddress(t *testing.T) {
	t.Parallel()

	c := &DHCPConfig{
		GatewayOffset: 1,
	}
	_, addr, _ := net.ParseCIDR("12.34.56.78/24")
	addr2 := c.GatewayAddress(addr)
	if addr2.String() != "12.34.56.1/24" {
		t.Error("wrong gateway address:", addr2.String())
	}
}

func testLeaseDuration(t *testing.T) {
	t.Parallel()

	c := &DHCPConfig{
		GatewayOffset: 100,
	}

	du := c.LeaseDuration()
	if du != DefaultLeaseDuration {
		t.Error(`du != DefaultLeaseDuration`)
	}

	c.LeaseMinutes = 30
	if c.LeaseDuration() != 30*time.Minute {
		t.Error(`c.LeaseDuration() != 30 * time.Minute`)
	}
}

func TestDHCP(t *testing.T) {
	t.Run("GatewayAddress", testGatewayAddress)
	t.Run("LeaseDuration", testLeaseDuration)
}
