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
prefix | Common prefix
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
`bmc`            | IP addresses of the machine's BMC indexed with protocol names (IPv4/IPv6)

`<prefix>/crypts/<serial>/<path>`
---------------------------------

Name   | Description
----   | -----------
prefix | Common prefix
serial | Serial number of a machine
path   | Name of an encrypted disk, in the format shown in `/dev/disk/by-path`

This type of key holds the encryption key of a disk.
The value is a raw binary key.

```console
$ etcdctl get /sabakan/crypts/1234abcd/pci-0000:00:1f.2-ata-3 --print-value-only
(This returns a binary key.)
```

`<prefix>/ipam`
---------------

Name   | Description
----   | -----------
prefix | Common prefix

This type of key holds IPAM configurations.
The value is [IPAMConfig](ipam.md#ipamconfig) formatted in JSON.

`<prefix>/node-indices/<rack>`
------------------------------

Name   | Description
----   | -----------
prefix | Common prefix
rack   | Rack nubmer

This type of key holds assignment of node indices per rack.
The value is a list of assigned indices formatted in JSON.

例:
```
$ etcdctl get "/sabakan/node-indices/0"
[3, 4, 5]
```
