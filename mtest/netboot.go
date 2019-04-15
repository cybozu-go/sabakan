package mtest

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// TestNetboot tests iPXE boot
func TestNetboot() {
	It("is achieved", func() {
		By("Set-up kernel params")
		sabactlSafe("kernel-params", "set", "\"coreos.autologin=ttyS0 console=ttyS0\"")

		By("Uploading an image")
		kernel := filepath.Join("/tmp", filepath.Base(coreosKernel))
		initrd := filepath.Join("/tmp", filepath.Base(coreosInitrd))
		sabactlSafe("images", "upload", coreosVersion, kernel, initrd)

		By("Waiting images to be distributed")
		Eventually(func() error {
			var index []struct {
				ID   string   `json:"id"`
				URLs []string `json:"urls"`
			}

			stdout, stderr, err := sabactl("images", "index")
			if err != nil {
				return fmt.Errorf("%v: stderr=%s", err, stderr)
			}
			err = json.Unmarshal([]byte(stdout), &index)
			if err != nil {
				return err
			}
			for _, img := range index {
				if img.ID != coreosVersion {
					continue
				}
				if len(img.URLs) == 3 {
					return errors.New("uploaded image does not have 3 urls")
				}
			}
			return nil
		}).Should(Succeed())

		By("Waiting worker to boot")
		Expect(prepareSSHClients(worker)).NotTo(HaveOccurred())

		By("Checking kernel boot parameter")
		stdout, stderr, err := execAt(worker, "cat", "/proc/cmdline")
		Expect(err).NotTo(HaveOccurred(), "stderr=%s", stderr)
		Expect(string(stdout)).To(ContainSubstring("coreos.autologin=ttyS0"))

		By("Checking encrypted disks")
		Eventually(func() error {
			_, stderr, err := execAt(worker, "ls", "/dev/mapper/crypt-*")
			if err != nil {
				return fmt.Errorf("%v: stderr=%s", err, stderr)
			}
			return nil
		}).Should(Succeed())

		By("Removing the image from the index")
		sabactlSafe("images", "delete", coreosVersion)

		By("Checking all servers remove the image")
		Eventually(func() error {
			for _, h := range []string{host1, host2, host3} {
				stdout, _, err := execAt(h, "ls", "/var/lib/sabakan/images/coreos")
				if err != nil || len(stdout) > 0 {
					return err
				}
			}
			return nil
		}).Should(Succeed())
	})
}
