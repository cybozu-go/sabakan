package dhcpd

import (
	"fmt"
	"sort"

	"go.universe.tf/netboot/dhcp4"
)

func getOptionsLog(pkt *dhcp4.Packet) map[string]interface{} {
	debugLog := make(map[string]interface{})
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
			out, err = pkt.Options.String(targetOpt)
			if err != nil {
				continue
			}
		}
		//fmt.Println(out)
		debugLog[fmt.Sprintf("option%d", n)] = out
	}
	return debugLog
}
