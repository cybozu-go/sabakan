package mtest

import (
	"os"
)

var (
	bridgeAddress = os.Getenv("BRIDGE_ADDRESS")
	host1         = os.Getenv("HOST1")
	host2         = os.Getenv("HOST2")
	host3         = os.Getenv("HOST3")
	worker        = os.Getenv("WORKER")

	coreosInitrd          = os.Getenv("COREOS_INITRD")
	coreosKernel          = os.Getenv("COREOS_KERNEL")
	coreosVersion         = os.Getenv("COREOS_VERSION")
	dhcpJSONPath          = os.Getenv("DHCP_JSON")
	etcdctlPath           = os.Getenv("ETCDCTL")
	ignitionsPath         = os.Getenv("IGNITIONS")
	ipamJSONPath          = os.Getenv("IPAM_JSON")
	machinesJSONPath      = os.Getenv("MACHINES_JSON")
	sabactlPath           = os.Getenv("SABACTL")
	sabakanPath           = os.Getenv("SABAKAN")
	sabakanCryptsetupPath = os.Getenv("SABAKAN_CRYPTSETUP")

	sshKeyFile = os.Getenv("SSH_PRIVKEY")
)
