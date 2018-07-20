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
	debug         = os.Getenv("DEBUG") == "1"
)
