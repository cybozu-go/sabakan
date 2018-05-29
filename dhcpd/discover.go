package dhcpd

import (
	"bytes"
	"context"
	"encoding/binary"
	"net"
	"strings"
	"time"

	"github.com/cybozu-go/log"
	"go.universe.tf/netboot/dhcp4"
)

func isUEFIHTTPBoot(pkt *dhcp4.Packet) bool {
	// RFC4578: Client System Architecture Type
	// Option 93 is a list of uint16 values
	bs, err := pkt.Options.Bytes(93)
	if err != nil {
		return false
	}

	if (len(bs) % 2) == 1 {
		return false
	}

	ok := false
	for i := 0; i < len(bs)/2; i++ {
		switch binary.BigEndian.Uint16(bs[i*2 : (i+1)*2]) {
		case 0x0F, 0x10:
			// x86/x64 UEFI HTTP Boot
			ok = true
		}
	}

	if !ok {
		return false
	}

	vcls, err := pkt.Options.String(dhcp4.OptVendorIdentifier)
	if err != nil {
		return false
	}

	return strings.HasPrefix(vcls, "HTTPClient")
}

func isIPXEBoot(pkt *dhcp4.Packet) bool {
	// RFC3004: User Class
	// Option 77 is a string.
	ucls, err := pkt.Options.String(77)
	if err != nil {
		return false
	}

	return ucls == "iPXE"
}

func isQemuMacAddress(mac net.HardwareAddr) bool {
	// "52:54:00" comes from placemat.
	return bytes.HasPrefix(mac, []byte{0x52, 0x54, 0x00})
}

func (h DHCPHandler) handleDiscover(ctx context.Context, pkt *dhcp4.Packet, intf Interface) (*dhcp4.Packet, error) {
	serverAddr, err := getIPv4AddrForInterface(intf)
	if err != nil {
		return nil, err
	}
	ifaddr := pkt.RelayAddr
	if ifaddr == nil || ifaddr.IsUnspecified() {
		ifaddr = serverAddr
	} else {
		// To delay answer to relayed requests, sleep shortly.
		time.Sleep(50 * time.Millisecond)
	}

	yourip, err := h.DHCP.Lease(ctx, ifaddr, pkt.HardwareAddr)
	if err != nil {
		return nil, err
	}
	opts, err := h.makeOptions(yourip)
	if err != nil {
		return nil, err
	}
	opts[dhcp4.OptServerIdentifier] = serverAddr
	resp := &dhcp4.Packet{
		Type:           dhcp4.MsgOffer,
		TransactionID:  pkt.TransactionID,
		Broadcast:      pkt.Broadcast,
		HardwareAddr:   pkt.HardwareAddr,
		YourAddr:       yourip,
		ServerAddr:     serverAddr,
		RelayAddr:      pkt.RelayAddr,
		BootServerName: serverAddr.String(),
		Options:        opts,
	}

	// UEFI HTTP Boot
	if isUEFIHTTPBoot(pkt) {
		log.Info("dhcp: requested UEFI HTTP boot", addPacketLog(pkt, map[string]interface{}{
			pktYiaddr: yourip.String(),
		}))
		opts[dhcp4.OptVendorIdentifier] = []byte("HTTPClient")
		resp.BootFilename = h.makeBootAPIURL(serverAddr, "ipxe.efi")
	}

	// iPXE Boot
	if isIPXEBoot(pkt) {
		log.Info("dhcp: requested iPXE boot", addPacketLog(pkt, map[string]interface{}{
			pktYiaddr: yourip.String(),
		}))
		// iPXE script to boot CoreOS Container Linux
		resp.BootFilename = h.makeBootAPIURL(serverAddr, "coreos/ipxe")
		if isQemuMacAddress(pkt.HardwareAddr) {
			resp.BootFilename += "?serial=1"
		}
	}

	return resp, nil
}
