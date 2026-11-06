package mtest

import (
	"encoding/json"
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
		waitAssetURLs("test")

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
		_, _, _ = etcdctl("compaction", "--physical=true", strconv.Itoa(currentRevision))

		By("Restarting host2 sabakan")
		_, _, err = startHost2Sabakan(sabakanImageURL)
		Expect(err).To(Succeed())

		By("Confirming that sabakan can get the latest asset again")
		waitAssetURLs("test2")
	})

	It("should restrict downloads of an asset with allowed-roles", func() {
		By("Uploading an asset with allowed-roles")
		execSafeAt(host1, "echo", "secret", ">", "restricted.txt")
		// --meta splits values by commas, so a single role is used here.
		sabactlSafe("assets", "upload", "--meta", "allowed-roles=worker", "restricted", "restricted.txt")

		By("Checking all servers pull the asset with the peer token")
		waitAssetURLs("restricted")

		By("Checking that clients that are not machines with the role are denied")
		for _, h := range []string{host1, host2, host3} {
			for _, server := range []string{host1, host2, host3} {
				stdout := execSafeAt(h, "curl", "-s", "-o", "/dev/null", "-w", "'%{http_code}'",
					"http://"+server+":10080/api/v1/assets/restricted")
				Expect(string(stdout)).To(Equal("403"), "client=%s server=%s", h, server)
			}
		}

		By("Checking that the metadata is still available")
		sabactlSafe("assets", "info", "restricted")

		By("Removing the asset")
		sabactlSafe("assets", "delete", "restricted")
	})
}

func waitAssetURLs(name string) {
	GinkgoHelper()
	Eventually(func() error {
		var info struct {
			Urls []string `json:"urls"`
		}

		stdout, stderr, err := sabactl("assets", "info", name)
		if err != nil {
			return fmt.Errorf("%v: stderr=%s", err, stderr)
		}
		if err := json.Unmarshal(stdout, &info); err != nil {
			return err
		}
		if len(info.Urls) != 3 {
			return fmt.Errorf("asset %s does not have 3 urls", name)
		}
		return nil
	}).Should(Succeed())
}
