package itest

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("images", func() {

	It("should be distributed between servers", func() {
		sabactl("images", "upload", coreosVersion, coreosKernel, coreosInitrd)

		Eventually(func() bool {
			var index []struct {
				ID   string   `json:"id"`
				URLs []string `json:"urls"`
			}

			stdout := sabactl("images", "index")
			err := json.Unmarshal([]byte(stdout), &index)
			if err != nil {
				return false
			}
			for _, img := range index {
				if img.ID != coreosVersion {
					continue
				}
				if len(img.URLs) == 3 {
					return true
				}
			}
			return false
		}).Should(BeTrue())
	})
})
