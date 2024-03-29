package mtest

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// testAssets tests sabakan assets
func testAssets() {
	It("should work as expected", func() {
		By("Uploading an asset")
		execSafeAt(host1, "echo", "test", ">", "asset.txt")
		sabactlSafe("assets", "upload", "test", "asset.txt")

		By("Checking all servers pull the asset")
		Eventually(func() error {
			var info struct {
				Urls []string `json:"urls"`
			}

			stdout, stderr, err := sabactl("assets", "info", "test")
			if err != nil {
				return fmt.Errorf("%v: stderr=%s", err, stderr)
			}
			err = json.Unmarshal(stdout, &info)
			if err != nil {
				return err
			}
			if len(info.Urls) != 3 {
				return errors.New("uploaded asset does not have 3 urls")
			}
			return nil
		}).Should(Succeed())

		By("Removing the asset from the index")
		sabactlSafe("assets", "delete", "test")

		By("Checking all servers remove the asset")
		Eventually(func() error {
			for _, h := range []string{host1, host2, host3} {
				stdout, _, err := execAt(h, "ls", "/var/lib/sabakan/assets")
				if err != nil || len(stdout) > 0 {
					return err
				}
			}
			return nil
		}).Should(Succeed())

		By("Stopping host2 sabakan")
		_, _, err := stopHost2Sabakan()
		Expect(err).To(Succeed())

		By("Adding two assets")
		execSafeAt(host1, "echo", "test1", ">", "update1.txt")
		sabactlSafe("assets", "upload", "test2", "update1.txt")
		execSafeAt(host1, "echo", "test2", ">", "update2.txt")
		sabactlSafe("assets", "upload", "test2", "update2.txt")

		By("Getting the current revision")
		stdout, stderr, err := etcdctl("get", "/", "-w=json")
		Expect(err).NotTo(HaveOccurred(), "stderr: %s", stderr)
		v := &struct {
			Header struct {
				Revision int `json:"revision"`
			} `json:"header"`
		}{}
		Expect(json.Unmarshal(stdout, v)).To(Succeed())
		currentRevision := v.Header.Revision

		By("Executing compaction")
		etcdctl("compaction", "--physical=true", strconv.Itoa(currentRevision))

		By("Restarting host2 sabakan")
		_, _, err = startHost2Sabakan(sabakanImageURL)
		Expect(err).To(Succeed())

		By("Confirming that sabakan can get the latest asset again")
		Eventually(func() error {
			var info struct {
				Urls []string `json:"urls"`
			}

			stdout, stderr, err := sabactl("assets", "info", "test2")
			if err != nil {
				return fmt.Errorf("%v: stderr=%s", err, stderr)
			}
			err = json.Unmarshal(stdout, &info)
			if err != nil {
				return err
			}

			if len(info.Urls) != 3 {
				return errors.New("uploaded asset does not have 3 urls")
			}
			return nil
		}).Should(Succeed())
	})
}
