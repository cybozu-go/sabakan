package mtest

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

		By("Set-up kernel params")
		sabactl("kernel-params", "set", "coreos.autologin=ttyS0")

		By("Waiting worker to boot")
		sshKey, err := parsePrivateKey()
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			c, err := sshTo(worker, sshKey)
			if err == nil {
				c.Close()
			}
			return err
		}).Should(Succeed())

		By("Waiting worker to boot")
		cli, err := sshTo(worker, sshKey)
		Expect(err).NotTo(HaveOccurred())
		defer cli.Close()

		sess, err := cli.NewSession()
		Expect(err).NotTo(HaveOccurred())
		var stdout, stderr bytes.Buffer
		sess.Stdout = &stdout
		sess.Stderr = &stderr
		err = sess.Run("cat /proc/cmdline")
		Expect(err).NotTo(HaveOccurred())
		sess.Close()

		Expect(stdout.String()).To(ContainSubstring("coreos.autologin=ttyS0"))

		By("Waiting worker to boot")
		By("Removing the image from the index")
		sabactl("images", "delete", coreosVersion)

		By("Checking all servers remove the image")
		Eventually(func() bool {
			for _, h := range []string{host1, host2, host3} {
				stdout, _, err := execAt(h, "ls", "/var/lib/sabakan/images/coreos")
				if err != nil || len(stdout) > 0 {
					return false
				}
			}
			return true
		}).Should(BeTrue())
	})
})
