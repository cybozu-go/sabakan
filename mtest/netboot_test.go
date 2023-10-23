package mtest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// testNetboot tests iPXE boot
func testNetboot() {
	It("is achieved", func() {
		By("Set-up kernel params")
		sabactlSafe("kernel-params", "set", "\"coreos.autologin=ttyS0 console=ttyS0\"")

		By("Uploading an image")
		kernel := filepath.Join("/var/tmp", filepath.Base(coreosKernel))
		initrd := filepath.Join("/var/tmp", filepath.Base(coreosInitrd))
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
		Expect(prepareSSHClients(worker1, worker2)).NotTo(HaveOccurred())

		for _, worker := range []string{worker1, worker2} {
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
				stdout, stderr, err = execAt(worker, "sudo", "dmsetup", "table", "/dev/mapper/crypt-*")
				if err != nil {
					return fmt.Errorf("%v: stderr=%s", err, stderr)
				}
				scanner := bufio.NewScanner(bytes.NewReader(stdout))
				for scanner.Scan() {
					hasNoReadWorkQueue := false
					hasNoWriteWorkQueue := false
					for _, field := range strings.Fields(scanner.Text()) {
						if field == "no_read_workqueue" {
							hasNoReadWorkQueue = true
						}
						if field == "no_write_workqueue" {
							hasNoWriteWorkQueue = true
						}
					}
					if !hasNoReadWorkQueue || !hasNoWriteWorkQueue {
						return fmt.Errorf("no_read_workqueue or no_write_workqueue is not set: %s", scanner.Text())
					}
				}
				if err = scanner.Err(); err != nil {
					return fmt.Errorf("reading stdout: %w", err)
				}
				return nil
			}, 6*time.Minute).Should(Succeed())

			By("Checking status of sabakan-cryptsetup")
			Eventually(func() error {
				stdout, _, err := execAt(worker, "systemctl", "is-active", "sabakan-cryptsetup")
				if err != nil {
					return err
				}
				if string(stdout) != "active\n" {
					return fmt.Errorf("sabakan-cryptsetup is not active:%s", stdout)
				}
				return nil
			}).Should(Succeed())
		}

		By("Copying readnvram binary")
		remoteFilename := filepath.Join("/var/tmp", filepath.Base(readNVRAM))
		copyReadNVRAM(worker2, remoteFilename)

		By("Reading encryption key from NVRAM")
		ekHexBefore, stderr, err := execAt(worker2, "sudo", remoteFilename)
		Expect(err).NotTo(HaveOccurred(), "stdout=%s, stderr=%s", ekHexBefore, stderr)

		By("Checking encryption key is kept after reboot")
		// Exit code is 255 when ssh is disconnected
		execAt(worker2, "sudo", "reboot")
		Expect(prepareSSHClients(worker2)).NotTo(HaveOccurred())
		copyReadNVRAM(worker2, remoteFilename)

		ekHexAfter, stderr, err := execAt(worker2, "sudo", remoteFilename)
		Expect(err).NotTo(HaveOccurred(), "stdout=%s, stderr=%s", ekHexAfter, stderr)
		Expect(ekHexAfter).To(Equal(ekHexBefore))

		By("Checking encrypted disks")
		Eventually(func() error {
			_, stderr, err := execAt(worker2, "ls", "/dev/mapper/crypt-*")
			if err != nil {
				return fmt.Errorf("%v: stderr=%s", err, stderr)
			}
			return nil
		}, 6*time.Minute).Should(Succeed())

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

func copyReadNVRAM(worker, remoteFilename string) {
	f, err := os.Open(readNVRAM)
	Expect(err).NotTo(HaveOccurred())
	defer f.Close()

	_, err = f.Seek(0, io.SeekStart)
	Expect(err).NotTo(HaveOccurred())
	stdout, stderr, err := execAtWithStream(worker, f, "dd", "of="+remoteFilename)
	Expect(err).NotTo(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
	stdout, stderr, err = execAt(worker, "chmod", "755", remoteFilename)
	Expect(err).NotTo(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
}
