package mtest

import (
	"encoding/json"
	"os"
	"strconv"

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

		By("Stopping host2 sabakan")
		Expect(stopHost2Sabakan()).To(Succeed())

		By("Adding two assets")
		f1 := localTempFile("updated 1")
		defer os.Remove(f1.Name())
		sabactl("assets", "upload", "test2", f1.Name())
		f2 := localTempFile("updated 2")
		defer os.Remove(f2.Name())
		sabactl("assets", "upload", "test2", f2.Name())

		By("Getting the current revision")
		stdout := etcdctl("get", "/", "-w=json")
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
		Expect(startHost2Sabakan()).To(Succeed())

		By("Confirming that sabakan can get the latest asset again")
		Eventually(func() bool {
			var info struct {
				Urls []string `json:"urls"`
			}

			stdout := sabactl("assets", "info", "test2")
			err := json.Unmarshal(stdout, &info)
			if err != nil {
				return false
			}
			return len(info.Urls) == 3
		}).Should(BeTrue())
	})
})

func stopHost2Sabakan() error {
	host2Client := sshClients[host2]
	sess, err := host2Client.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()
	return sess.Run("sudo systemctl stop sabakan.service")
}

func startHost2Sabakan() error {
	host2Client := sshClients[host2]
	sess, err := host2Client.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()
	return sess.Run("sudo systemd-run --unit=sabakan.service /data/sabakan -config-file /etc/sabakan.yml")
}
