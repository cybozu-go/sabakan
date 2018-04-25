package dhcp4

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/cybozu-go/log"
	"go.universe.tf/netboot/dhcp4"
)

// Server is a DHCP server
type Server interface {
	Serve(ctx context.Context) error
}

// New creates a new dhcp Server object
func New(bind string, ifname string, ipxe string, begin, end net.IP) Server {
	s := new(dhcpserver)
	s.bind = bind
	s.ifname = ifname
	s.ipxe = ipxe
	s.assign = assignment{
		begin:  begin,
		end:    end,
		leases: make(map[uint32]struct{}),
	}
	return s
}

type dhcpserver struct {
	bind   string
	ifname string
	ipxe   string

	assign assignment
}

// Architecture represent an architecture type
type Architecture int

// Architecture type
const (
	ArchIA32 Architecture = iota
	ArchX64
)

// Firmware represent a firmware type
type Firmware int

// Firmware type
const (
	FirmwareX86PC Firmware = iota
	FirmwareEFI32
	FirmwareEFI64
	FirmwareEFIBC
	FirmwareX86Ipxe
)

type dhcpConn interface {
	Close() error
	RecvDHCP() (*dhcp4.Packet, *net.Interface, error)
	SendDHCP(pkt *dhcp4.Packet, intf *net.Interface) error
}

func (s *dhcpserver) handleDiscover(conn dhcpConn, pkt *dhcp4.Packet, intf *net.Interface) error {
	/*
		if err = s.isBootDHCP(pkt); err != nil {
			log.Debug("DHCP: Ignoring packet", map[string]interface{}{
				"mac_address": pkt.HardwareAddr,
				"error":       err,
			})
			continue
		}
	*/
	arch, fwtype, err := s.validateDHCP(pkt)
	/*
		if err != nil {
			log.Debug("DHCP: Unusable packet", map[string]interface{}{
				"mac_address": pkt.HardwareAddr,
				"error":       err,
			})
			continue
		}
	*/

	log.Debug("DHCP: Got valid request to boot", map[string]interface{}{
		"mac_address": pkt.HardwareAddr,
		"error":       err,
	})
	log.Debug("DHCP: Got valid request", map[string]interface{}{
		"mac_address":  pkt.HardwareAddr,
		"architecture": arch,
	})

	ip, err := s.assign.next()
	if err != nil {
		log.Info("DHCP: Couldn't allocate ip address", map[string]interface{}{
			"mac_address": pkt.HardwareAddr,
			"error":       err,
		})
		return nil
	}

	serverIP, err := interfaceIP(intf)
	if err != nil {
		log.Info("DHCP: Couldn't get a source address", map[string]interface{}{
			"mac_address": pkt.HardwareAddr,
			"interface":   intf.Name,
			"error":       err,
		})
		return nil
	}

	resp, err := s.offerDHCP(pkt, serverIP, arch, fwtype, ip)
	if err != nil {
		log.Info("DHCP: Failed to construct ProxyDHCP offer", map[string]interface{}{
			"mac_address": pkt.HardwareAddr,
			"error":       err,
		})
		return nil
	}

	if err = conn.SendDHCP(resp, intf); err != nil {
		log.Info("DHCP: Failed to send ProxyDHCP offer", map[string]interface{}{
			"mac_address": pkt.HardwareAddr,
			"error":       err,
		})
		return nil
	}

	return nil
}

func (s *dhcpserver) handleRequest(conn dhcpConn, pkt *dhcp4.Packet, intf *net.Interface) error {

	ip := pkt.Options[dhcp4.OptRequestedIP]

	serverIP, err := interfaceIP(intf)

	resp, err := s.ackDHCP(pkt, serverIP, ip)
	if err != nil {
		log.Info("DHCP: Failed to construct ProxyDHCP ack", map[string]interface{}{
			"mac_address": pkt.HardwareAddr,
			"error":       err,
		})
		return nil
	}

	if err = conn.SendDHCP(resp, intf); err != nil {
		log.Info("DHCP: Failed to send ProxyDHCP ack", map[string]interface{}{
			"mac_address": pkt.HardwareAddr,
			"error":       err,
		})
		return nil
	}
	return nil
}

func (s *dhcpserver) Serve(ctx context.Context) error {
	conn, err := dhcp4.NewConn(s.bind)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	for {
		pkt, intf, err := conn.RecvDHCP()
		if err != nil {
			return fmt.Errorf("Receiving DHCP packet: %s", err)
		}
		/*
			if intf.Name != s.ifname {
				log.Debug("DHCP: Ignoring packet", map[string]interface{}{
					"listen_interface": s.ifname,
					"received_on":      intf.Name,
				})
				continue
			}
		*/

		switch pkt.Type {
		case dhcp4.MsgDiscover:
			err = s.handleDiscover(conn, pkt, intf)
		case dhcp4.MsgRequest:
			err = s.handleRequest(conn, pkt, intf)
		default:
			err = fmt.Errorf("unknown packet type: %v", pkt.Type)
		}

		if err != nil {
			return err
		}

	}
}

func (s *dhcpserver) isBootDHCP(pkt *dhcp4.Packet) error {
	if pkt.Type != dhcp4.MsgDiscover {
		return fmt.Errorf("packet is %s, not %s", pkt.Type, dhcp4.MsgDiscover)
	}

	if pkt.Options[93] == nil {
		return errors.New("not a PXE boot request (missing option 93)")
	}

	return nil
}

func (s *dhcpserver) validateDHCP(pkt *dhcp4.Packet) (arch Architecture, fwtype Firmware, err error) {
	fwt, err := pkt.Options.Uint16(93)
	if err != nil {
		return 0, 0, fmt.Errorf("malformed DHCP option 93 (required for PXE): %s", err)
	}

	switch fwt {
	case 0:
		arch = ArchIA32
		fwtype = FirmwareX86PC
	case 6:
		arch = ArchIA32
		fwtype = FirmwareEFI32
	case 7:
		arch = ArchX64
		fwtype = FirmwareEFI64
	case 9:
		arch = ArchX64
		fwtype = FirmwareEFIBC
	default:
		return 0, 0, fmt.Errorf("unsupported client firmware type '%d'", fwtype)
	}

	if class, err := pkt.Options.String(77); err == nil {
		if class == "iPXE" && fwtype == FirmwareX86PC {
			fwtype = FirmwareX86Ipxe
		}
	}
	return arch, fwtype, nil
}

func (s *dhcpserver) offerDHCP(pkt *dhcp4.Packet, serverIP net.IP, arch Architecture, fwtype Firmware, clientIP net.IP) (*dhcp4.Packet, error) {
	resp := &dhcp4.Packet{
		Type:          dhcp4.MsgOffer,
		TransactionID: pkt.TransactionID,
		Broadcast:     true,
		HardwareAddr:  pkt.HardwareAddr,
		RelayAddr:     pkt.RelayAddr,
		ServerAddr:    serverIP,
		ClientAddr:    nil,
		YourAddr:      clientIP,
		Options:       make(dhcp4.Options),
	}

	switch fwtype {
	case FirmwareEFI32, FirmwareEFI64, FirmwareEFIBC:
		resp.BootServerName = serverIP.String()
		resp.BootFilename = s.ipxe
		resp.Options[dhcp4.OptServerIdentifier] = serverIP
		resp.Options[dhcp4.OptVendorIdentifier] = []byte("HTTPClient")
		resp.Options[97] = pkt.Options[97]
	default:
		resp.Options[dhcp4.OptDHCPMessageType] = []byte{2}
	}

	return resp, nil
}

func (s *dhcpserver) ackDHCP(pkt *dhcp4.Packet, serverIP net.IP, clientIP net.IP) (*dhcp4.Packet, error) {
	resp := &dhcp4.Packet{
		Type:          dhcp4.MsgAck,
		TransactionID: pkt.TransactionID,
		Broadcast:     true,
		HardwareAddr:  pkt.HardwareAddr,
		RelayAddr:     pkt.RelayAddr,
		ServerAddr:    serverIP,
		ClientAddr:    nil,
		YourAddr:      clientIP,
		Options:       make(dhcp4.Options),
	}

	resp.Options[dhcp4.OptDHCPMessageType] = []byte{5}

	return resp, nil
}

func interfaceIP(intf *net.Interface) (net.IP, error) {
	addrs, err := intf.Addrs()
	if err != nil {
		return nil, err
	}

	// Try to find an IPv4 address to use, in the following order:
	// global unicast (includes rfc1918), link-local unicast,
	// loopback.
	fs := [](func(net.IP) bool){
		net.IP.IsGlobalUnicast,
		net.IP.IsLinkLocalUnicast,
		net.IP.IsLoopback,
	}
	for _, f := range fs {
		for _, a := range addrs {
			ipaddr, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipaddr.IP.To4()
			if ip == nil {
				continue
			}
			if f(ip) {
				return ip, nil
			}
		}
	}

	return nil, errors.New("no usable unicast address configured on interface")
}
