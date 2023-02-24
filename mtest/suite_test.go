package mtest

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMtest(t *testing.T) {
	if os.Getenv("SSH_PRIVKEY") == "" {
		t.Skip("no SSH_PRIVKEY envvar")
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "Multi-host test for sabakan")
}

var _ = BeforeSuite(func() {
	RunBeforeSuite()
})

// This must be the only top-level test container.
// Other tests and test containers must be listed in this.
var _ = Describe("Test Sabakan functions", func() {
	Context("assets", testAssets)
	Context("netboot", testNetboot)
})
