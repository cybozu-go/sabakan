package itest

import (
	"os"
)

var (
	bridgeAddress = os.Getenv("BRIDGE_ADDRESS")
	host1         = os.Getenv("HOST1")
	host2         = os.Getenv("HOST2")
	host3         = os.Getenv("HOST3")
	sshKeyFile    = os.Getenv("SSH_PRIVKEY")
	sabactlPath   = os.Getenv("SABACTL")
	coreosVersion = os.Getenv("COREOS_VERSION")
	coreosKernel  = os.Getenv("COREOS_KERNEL")
	coreosInitrd  = os.Getenv("COREOS_INITRD")
	debug         = os.Getenv("DEBUG") == "1"
)
