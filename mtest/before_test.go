package mtest

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// RunBeforeSuite is for Ginkgo BeforeSuite
func RunBeforeSuite() {
	fmt.Println("Preparing...")

	SetDefaultEventuallyPollingInterval(time.Second)
	SetDefaultEventuallyTimeout(3 * time.Minute)

	err := prepareSSHClients(host1, host2, host3)
	Expect(err).NotTo(HaveOccurred())

	// sync VM root filesystem to store newly generated SSH host keys.
	for h := range sshClients {
		execSafeAt(h, "sync")
	}

	By("copying test files")
	for _, testFile := range []string{etcdPath, etcdctlPath} {
		f, err := os.Open(testFile)
		Expect(err).NotTo(HaveOccurred())
		defer f.Close()
		remoteFilename := filepath.Join("/var/tmp", filepath.Base(testFile))
		for _, host := range []string{host1, host2, host3} {
			_, err := f.Seek(0, io.SeekStart)
			Expect(err).NotTo(HaveOccurred())
			stdout, stderr, err := execAtWithStream(host, f, "dd", "of="+remoteFilename)
			Expect(err).NotTo(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
			stdout, stderr, err = execAt(host, "sudo", "mv", remoteFilename, filepath.Join("/opt/bin", filepath.Base(testFile)))
			Expect(err).NotTo(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
			stdout, stderr, err = execAt(host, "sudo", "chmod", "755", filepath.Join("/opt/bin", filepath.Base(testFile)))
			Expect(err).NotTo(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		}
	}
	for _, testFile := range []string{coreosKernel, coreosInitrd, sshKeyFile, sabakanImagePath} {
		f, err := os.Open(testFile)
		Expect(err).NotTo(HaveOccurred())
		defer f.Close()
		remoteFilename := filepath.Join("/var/tmp", filepath.Base(testFile))
		for _, host := range []string{host1, host2, host3} {
			_, err := f.Seek(0, io.SeekStart)
			Expect(err).NotTo(HaveOccurred())
			stdout, stderr, err := execAtWithStream(host, f, "dd", "of="+remoteFilename)
			Expect(err).NotTo(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		}
	}

	By("loading test image")
	err = loadImage(filepath.Join("/var/tmp", filepath.Base(sabakanImagePath)))
	Expect(err).NotTo(HaveOccurred())

	By("running install-tools")
	err = installTools(sabakanImageURL)
	Expect(err).NotTo(HaveOccurred())

	By("starting etcd")
	err = stopEtcd()
	Expect(err).NotTo(HaveOccurred())
	err = runEtcd()
	Expect(err).NotTo(HaveOccurred())

	time.Sleep(time.Second)

	By("starting sabakan")
	err = stopSabakan()
	Expect(err).NotTo(HaveOccurred())
	err = runSabakan(sabakanImageURL)
	Expect(err).NotTo(HaveOccurred())

	// wait sabakan
	Eventually(func() error {
		_, _, err := sabactl("logs")
		return err
	}).Should(Succeed())

	By("configuring sabakan")
	// register ipam.json, dhcp.json, machines.json, and ignitions
	ipam, err := os.ReadFile(ipamJSONPath)
	Expect(err).NotTo(HaveOccurred())
	ipamFile := remoteTempFile(string(ipam))
	sabactlSafe("ipam", "set", "-f", ipamFile)

	dhcp, err := os.ReadFile(dhcpJSONPath)
	Expect(err).NotTo(HaveOccurred())
	dhcpFile := remoteTempFile(string(dhcp))
	sabactlSafe("dhcp", "set", "-f", dhcpFile)

	sabactlSafe("ignitions", "set", "-f", "/ignitions/worker.yml", "worker", "1.0.0")

	machines, err := os.ReadFile(machinesJSONPath)
	Expect(err).NotTo(HaveOccurred())
	machinesFile := remoteTempFile(string(machines))
	sabactlSafe("machines", "create", "-f", machinesFile)

	time.Sleep(time.Second)
	fmt.Println("Begin tests...")
}
