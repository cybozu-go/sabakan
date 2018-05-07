package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"path"

	"github.com/cybozu-go/netutil"
	"github.com/cybozu-go/sabakan"
)

func offsetToInt(offset string) uint32 {
	ip, _, _ := net.ParseCIDR(offset)
	return netutil.IP4ToInt(ip)
}

func (d *Driver) generateIP(ctx context.Context, mc *sabakan.Machine) (*sabakan.MachineJSON, error) {
	/*
		Generate IP addresses by sabakan config
		Example:
			net0 = node-ipv4-offset + (1 << node-rack-shift * 1 * rack-number) + node-number-of-a-rack
			net1 = node-ipv4-offset + (1 << node-rack-shift * 2 * rack-number) + node-number-of-a-rack
			net2 = node-ipv4-offset + (1 << node-rack-shift * 3 * rack-number) + node-number-of-a-rack
			bmc  = bmc-ipv4-offset + (1 << bmc-rack-shift * 1 * rack-number) + node-number-of-a-rack
	*/
	res := mc.ToJSON()
	key := path.Join(d.prefix, KeyConfig)
	resp, err := d.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, errors.New("ipam config is not found")
	}

	var sc sabakan.IPAMConfig
	err = json.Unmarshal(resp.Kvs[0].Value, &sc)
	if err != nil {
		return nil, err
	}

	res.Network = map[string]sabakan.MachineNetwork{}
	for i := 0; i < int(sc.NodeIPPerNode); i++ {
		uintip := offsetToInt(sc.NodeIPv4Offset) + (uint32(1) << uint32(sc.NodeRackShift) * uint32(i+1) * mc.Rack) + mc.NodeNumberOfRack
		ip := netutil.IntToIP4(uintip)
		ifname := fmt.Sprintf("net%d", i)
		res.Network[ifname] = sabakan.MachineNetwork{
			IPv4: []string{ip.String()},
			IPv6: []string{},
			Mac:  "",
		}
	}
	for i := 0; i < int(sc.BMCIPPerNode); i++ {
		uintip := offsetToInt(sc.BMCIPv4Offset) + (uint32(1) << uint32(sc.BMCRackShift) * uint32(i+1) * mc.Rack) + mc.NodeNumberOfRack
		ip := netutil.IntToIP4(uintip)
		res.BMC = sabakan.MachineBMC{
			IPv4: []string{ip.String()},
		}
	}

	return res, nil
}
