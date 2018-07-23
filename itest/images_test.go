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
				Id   string   `json:"id"`
				Urls []string `json:"urls"`
			}

			stdout := sabactl("images", "index")
			err := json.Unmarshal([]byte(stdout), &index)
			if err != nil {
				return false
			}
			for _, img := range index {
				if img.Id != "id" {
					continue
				}
				if len(img.Urls) == 3 {
					return true
				}
			}
			return false
		}).Should(BeTrue())

		sabactl("images", "delete", "id")
	})
})
