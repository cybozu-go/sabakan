package dhcpd

import "net"

// Interface is an abstract network interface.
type Interface interface {
	Addrs() ([]net.Addr, error)
	Name() string
}

type nativeInterface struct {
	*net.Interface
}

func (i nativeInterface) Name() string {
	return i.Interface.Name
}
