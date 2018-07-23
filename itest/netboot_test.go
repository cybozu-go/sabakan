package itest

import "encoding/json"

var _ = Describe("worker", func() {

	It("should successfully boots via HTTP/iPXE", func() {

		By("Waiting CoreOS images to be uploaded")
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

		By("Waiting worker to boot")
	})
})
