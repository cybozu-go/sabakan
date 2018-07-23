package itest

import (
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestItest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration test for sabakan")
}

var _ = BeforeSuite(func() {
	fmt.Println("Preparing...")

	SetDefaultEventuallyPollingInterval(time.Second)
	SetDefaultEventuallyTimeout(time.Minute)

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
	}).Should(Succeed(), 3*time.Minute)

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

	time.Sleep(time.Second)
	fmt.Println("Begin tests...")
})
