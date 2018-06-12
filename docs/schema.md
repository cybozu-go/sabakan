Data Schema in etcd
===================

Sabakan stores various types of data in etcd.
Keys and values in etcd are described below.

All keys are prefixed with a string specified in the sabakan command-line option.
This prefix string is denoted as `<prefix>` in the following.

`<prefix>/machines/<serial>`
----------------------------

Name   | Description
----   | -----------
serial | Serial number of a machine

This type of key holds the information of a machine.
The value is formatted in JSON.

```console
$ etcdctl get /sabakan/machines/1234abcd --print-value-only | jq .
{
  "serial": "1234abcd",
  "product": "R630",
  "datacenter": "ty3",
  "rack": 1,
  "index-in-rack": 1,
  "role": "boot",
  "network": {
    "net0": {
      "ipv4": [
        "10.69.0.69"
      ],
      "ipv6": []
    },
    "net1": {
      "ipv4": [
        "10.69.0.133"
      ],
      "ipv6": []
    }
  },
  "bmc": {
    "type": "iDRAC-9",
    "ipv4": "10.72.17.37"
  }
```

Key              | Description
---              | -----------
`serial`         | Serial number of the machine
`product`        | Product name of the machine
`datacenter`     | Data center name where the machine is in
`rack`           | Logical rack number (LRN) where the machine is in
`index-in-rack`  | Index number of the machine in a rack; this does not correspond to physical position
`network`        | IP addresses of the machine indexed with NIC names and protocol names (IPv4/IPv6)
`bmc`            | Machine's BMC specs; see below

Key in `bmc`    | Description
------------    | -----------
`type`          | BMC type e.g. 'iDRAC-9', 'IPMI-2.0'
`ipv4`          | IPv4 address of BMC
`ipv6`          | IPv6 address of BMC

`<prefix>/crypts/<serial>/<path>`
---------------------------------

Name   | Description
----   | -----------
serial | Serial number of a machine
path   | Name of an encrypted disk, in the format shown in `/dev/disk/by-path`

This type of key holds the encryption key of a disk.
The value is a raw binary key.

```console
$ etcdctl get /sabakan/crypts/1234abcd/pci-0000:00:1f.2-ata-3 --print-value-only
(This returns a binary key.)
```

`<prefix>/images/coreos`
------------------------

This type of key holds the index of CoreOS container Linux images.
The value is described in [boot image management](image_management.md).

`<prefix>/images/coreos/deleted`
--------------------------------

This type of key holds a list of deleted image IDs as follows:

```javascript
["123.45.6", "789.0.1", "2018.04.01"]
```

`<prefix>/assets/`
------------------

This key stores the last ID of the asset.  Each time an asset is added
or updated, the value will be incremented by one.

`<prefix>/assets/<NAME>`
------------------------

This type of key holds the meta data of an asset.
The value is described in [asset management](assets.md).

`<prefix>/ignitions/<role>/<id>`
-------------------------

This type of key holds an ignition template. Ignitions are distinguished by `<role>`.

`<prefix>/ipam`
---------------

This type of key holds IPAM configurations.
The value is [IPAMConfig](ipam.md#ipamconfig) formatted in JSON.

`<prefix>/dhcp`
---------------

This type of key holds DHCP configurations.
The value is [DHCPConfig](dhcp.md#dhcpconfig) formatted in JSON.

`<prefix>/lease-usages/<ip>`
----------------------------

Name | Description
---- | -----------
ip   | The first IP address of the lease range.

These keys hold lease address usages for a range of IP addresses.
The value is a mapping between hardware address and (`index`, `expire`)
pair where `index` is the index of the leased IP address in the range
and `expire` is the Go's `time.Time` when the lease expires.

`<prefix>/node-indices/<rack>`
------------------------------

Name | Description
---- | -----------
rack | Rack nubmer

This type of key holds assignment of node indices per rack.
The value is a list of assigned indices formatted in JSON.

例:
```
$ etcdctl get "/sabakan/node-indices/0"
[3, 4, 5]
```
