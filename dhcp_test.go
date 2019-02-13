package sabakan

import (
	"testing"
	"time"
)

func testLeaseDuration(t *testing.T) {
	t.Parallel()

	c := &DHCPConfig{}

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
	t.Run("LeaseDuration", testLeaseDuration)
}
