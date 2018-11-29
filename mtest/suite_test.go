package mtest

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestItest(t *testing.T) {
	if len(sshKeyFile) == 0 {
		t.Skip("no SSH_PRIVKEY envvar")
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration test for sabakan")
}

var _ = BeforeSuite(func() {
	fmt.Println("Preparing...")

	SetDefaultEventuallyPollingInterval(time.Second)
	SetDefaultEventuallyTimeout(3 * time.Minute)

	err := prepareSSHClients(host1, host2, host3)
	Expect(err).NotTo(HaveOccurred())

	Eventually(func() error {
		for _, host := range []string{host1, host2, host3} {
			_, _, err := execAt(host, "test -f /boot/ipxe.efi")
			if err != nil {
				return err
			}
		}
		return nil
	}).Should(Succeed())

	// sync VM root filesystem to store newly generated SSH host keys.
	for h := range sshClients {
		execSafeAt(h, "sync")
	}

	err = stopEtcd(sshClients[host1])
	Expect(err).NotTo(HaveOccurred())
	err = runEtcd(sshClients[host1])
	Expect(err).NotTo(HaveOccurred())

	time.Sleep(time.Second)

	err = stopSabakan()
	Expect(err).NotTo(HaveOccurred())
	err = runSabakan()
	Expect(err).NotTo(HaveOccurred())

	// wait sabakan
	Eventually(func() error {
		_, _, err := execAt(host1, "/data/sabactl", "logs")
		return err
	}).Should(Succeed())

	// register ipam.json, dhcp.json, machines.json, and ignitions
	sabactl("ipam", "set", "-f", ipamJSONPath)
	sabactl("dhcp", "set", "-f", dhcpJSONPath)
	sabactl("machines", "create", "-f", machinesJSONPath)
	ignitionSource := filepath.Join(ignitionsPath, "worker.yml")
	sabactl("ignitions", "set", "-f", ignitionSource, "worker", "1.0.0")

	time.Sleep(time.Second)
	fmt.Println("Begin tests...")
})
