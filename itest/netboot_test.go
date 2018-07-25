package itest

import (
	"bytes"
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("netboot", func() {

	It("is achieved", func() {
		By("Uploading an image")
		sabactl("images", "upload", coreosVersion, coreosKernel, coreosInitrd)

		By("Waiting images to be distributed")
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

		By("Waiting worker to boot")
		sshKey, err := parsePrivateKey()
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			_, err := sshTo(worker, sshKey)
			return err
		}).Should(Succeed())

		sabactl("images", "delete", coreosVersion)
		stdout := sabactl("images", "index")
		var images []string
		err = json.NewDecoder(bytes.NewReader(stdout)).Decode(&images)
		if err != nil {
			Fail(err.Error())
		}
		Expect(images).NotTo(ContainElement(coreosVersion))
	})
})
