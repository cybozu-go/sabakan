package itest

import (
	"encoding/json"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("assets", func() {

	It("should work as expected", func() {
		f := localTempFile("test")
		defer os.Remove(f.Name())

		By("Uploading an asset")
		sabactl("assets", "upload", "test", f.Name())

		By("Checking all servers pull the asset")
		Eventually(func() bool {
			var info struct {
				Urls []string `json:"urls"`
			}

			stdout := sabactl("assets", "info", "test")
			err := json.Unmarshal(stdout, &info)
			if err != nil {
				return false
			}
			return len(info.Urls) == 3
		}).Should(BeTrue())

		By("Removing the asset from the index")
		sabactl("assets", "delete", "test")

		By("Checking all servers remove the asset")
		Eventually(func() bool {
			for _, h := range []string{host1, host2, host3} {
				stdout, _, err := execAt(h, "ls", "/var/lib/sabakan/assets")
				if err != nil || len(stdout) > 0 {
					return false
				}
			}
			return true
		}).Should(BeTrue())
	})
})
