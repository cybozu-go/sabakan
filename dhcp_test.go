package sabakan

import (
	"net"
	"testing"
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

func TestDHCP(t *testing.T) {
	t.Run("GatewayAddress", testGatewayAddress)
}
