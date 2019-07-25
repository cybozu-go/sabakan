package mtest

import (
	"os"
)

var (
	bridgeAddress = os.Getenv("BRIDGE_ADDRESS")
	host1         = os.Getenv("HOST1")
	host2         = os.Getenv("HOST2")
	host3         = os.Getenv("HOST3")
	worker1       = os.Getenv("WORKER1")
	worker2       = os.Getenv("WORKER2")

	coreosInitrd     = os.Getenv("COREOS_INITRD")
	coreosKernel     = os.Getenv("COREOS_KERNEL")
	coreosVersion    = os.Getenv("COREOS_VERSION")
	dhcpJSONPath     = os.Getenv("DHCP_JSON")
	etcdPath         = os.Getenv("ETCD")
	etcdctlPath      = os.Getenv("ETCDCTL")
	ignitionsPath    = os.Getenv("IGNITIONS")
	ipamJSONPath     = os.Getenv("IPAM_JSON")
	machinesJSONPath = os.Getenv("MACHINES_JSON")
	sabakanImagePath = os.Getenv("SABAKAN_IMAGE")
	sabakanImageURL  = os.Getenv("SABAKAN_IMAGE_URL")

	sshKeyFile = os.Getenv("SSH_PRIVKEY")

	readNVRAM = os.Getenv("READ_NVRAM")
)
