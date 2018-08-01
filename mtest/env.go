package mtest

import (
	"os"
)

var (
	bridgeAddress    = os.Getenv("BRIDGE_ADDRESS")
	host1            = os.Getenv("HOST1")
	host2            = os.Getenv("HOST2")
	host3            = os.Getenv("HOST3")
	worker           = os.Getenv("WORKER")
	sshKeyFile       = os.Getenv("SSH_PRIVKEY")
	sabactlPath      = os.Getenv("SABACTL")
	etcdctlPath      = os.Getenv("ETCDCTL")
	ipamJSONPath     = os.Getenv("IPAM_JSON")
	dhcpJSONPath     = os.Getenv("DHCP_JSON")
	machinesJSONPath = os.Getenv("MACHINES_JSON")
	ignitionsPath    = os.Getenv("IGNITIONS")
	coreosVersion    = os.Getenv("COREOS_VERSION")
	coreosKernel     = os.Getenv("COREOS_KERNEL")
	coreosInitrd     = os.Getenv("COREOS_INITRD")
	debug            = os.Getenv("DEBUG") == "1"
)
