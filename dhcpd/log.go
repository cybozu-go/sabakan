package dhcpd

import (
	"encoding/binary"
	"fmt"
	"net"
	"sort"

	"go.universe.tf/netboot/dhcp4"
)

const (
	pktCiaddr = "ciaddr"
	pktYiaddr = "yiaddr"
)

var optionNames = map[dhcp4.Option]string{
	dhcp4.OptSubnetMask:         "subnet_mask",
	dhcp4.OptTimeOffset:         "time_offset",
	dhcp4.OptRouters:            "router",
	dhcp4.OptDNSServers:         "domain_name_server",
	dhcp4.OptHostname:           "host_name",
	dhcp4.OptBootFileSize:       "boot_file_size",
	dhcp4.OptDomainName:         "domain_name",
	dhcp4.OptBroadcastAddr:      "broadcast_address",
	dhcp4.OptNTPServers:         "ntp_server",
	dhcp4.OptVendorSpecific:     "vender_specific",
	dhcp4.OptRequestedIP:        "requested_ip_address",
	dhcp4.OptLeaseTime:          "lease_time",
	dhcp4.OptOverload:           "overload",
	dhcp4.OptServerIdentifier:   "server_identifier",
	dhcp4.OptRequestedOptions:   "requested_options",
	dhcp4.OptMessage:            "message",
	dhcp4.OptMaximumMessageSize: "maximum_message_size",
	dhcp4.OptRenewalTime:        "renewal_time",
	dhcp4.OptRebindingTime:      "rebinding_time",
	dhcp4.OptVendorIdentifier:   "vendor_class_identifier",
	dhcp4.OptClientIdentifier:   "client_identifier",
	dhcp4.OptFQDN:               "fqdn",
}

func optionLogKey(n dhcp4.Option) string {
	return fmt.Sprintf("option_%d_%s", n, optionNames[n])
}

func getPacketLog(pkt *dhcp4.Packet, intf *net.Interface) map[string]interface{} {
	pktLog := map[string]interface{}{
		"intf":      intf.Name,
		"type":      pkt.Type.String(),
		"xid":       binary.BigEndian.Uint32(pkt.TransactionID),
		"broadcast": pkt.Broadcast,
		"chaddr":    pkt.HardwareAddr,
	}
	if len(pkt.ClientAddr) > 0 && !pkt.ClientAddr.Equal(net.IPv4zero) {
		pktLog["ciaddr"] = pkt.ClientAddr
	}
	if len(pkt.YourAddr) > 0 && !pkt.YourAddr.Equal(net.IPv4zero) {
		pktLog["yiaddr"] = pkt.YourAddr
	}
	if len(pkt.ServerAddr) > 0 && !pkt.ServerAddr.Equal(net.IPv4zero) {
		pktLog["siaddr"] = pkt.ServerAddr
	}
	if len(pkt.RelayAddr) > 0 && !pkt.RelayAddr.Equal(net.IPv4zero) {
		pktLog["giaddr"] = pkt.RelayAddr
	}
	if len(pkt.BootServerName) > 0 {
		pktLog["sname"] = pkt.BootServerName
	}

	return pktLog
}

func getOptionsLog(pkt *dhcp4.Packet) map[string]interface{} {
	optLog := make(map[string]interface{})

	optLog["xid"] = binary.BigEndian.Uint32(pkt.TransactionID)

	var opts []int
	for n := range pkt.Options {
		opts = append(opts, int(n))
	}
	sort.Ints(opts)
	for _, n := range opts {
		targetOpt := dhcp4.Option(n)
		var out interface{}
		var err error
		switch targetOpt {
		case dhcp4.OptSubnetMask:
			mask, err := pkt.Options.IPMask(targetOpt)
			if err != nil {
				continue
			}
			ones, _ := mask.Size()
			out = fmt.Sprintf("/%d", ones)
		case dhcp4.OptBroadcastAddr, dhcp4.OptNTPServers, dhcp4.OptServerIdentifier:
			out, err = pkt.Options.IP(targetOpt)
			if err != nil {
				continue
			}
		case dhcp4.OptRouters, dhcp4.OptDNSServers:
			out, err = pkt.Options.IPs(targetOpt)
			if err != nil {
				continue
			}
		case dhcp4.OptLeaseTime, dhcp4.OptRenewalTime, dhcp4.OptRebindingTime:
			out, err = pkt.Options.Uint32(targetOpt)
			if err != nil {
				continue
			}
		case dhcp4.OptTimeOffset:
			out, err = pkt.Options.Int32(targetOpt)
			if err != nil {
				continue
			}
		case dhcp4.OptBootFileSize, dhcp4.OptMaximumMessageSize:
			out, err = pkt.Options.Uint16(targetOpt)
			if err != nil {
				continue
			}
		default:
			// TODO: escape non-ASCII string
			out, err = pkt.Options.String(targetOpt)
			if err != nil {
				continue
			}
		}
		optLog[optionLogKey(targetOpt)] = out
	}
	return optLog
}

func addPacketLog(pkt *dhcp4.Packet, fields map[string]interface{}) map[string]interface{} {
	ret := fields
	if ret == nil {
		ret = make(map[string]interface{})
	}
	ret["xid"] = binary.BigEndian.Uint32(pkt.TransactionID)
	ret["chaddr"] = pkt.HardwareAddr.String()
	return ret
}
