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
The value is formatted in JSON as defined in [Machine](machine.md).

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

`<prefix>/assets`
-----------------

This key stores the last ID of the asset.  Each time an asset is added
or updated, the value will be incremented by one.

`<prefix>/assets/<NAME>`
------------------------

This type of key holds the meta data of an asset.
The value is described in [asset management](assets.md).

`<prefix>/ignitions/<role>/<id>`
-------------------------

This type of key holds an ignition template. Ignitions are distinguished by `<role>`.

`<prefix>/ignitions/<role>/<id>/meta`
-------------------------------------

This type of key holds the meta data of an ignition template.
The value is `map[string]string` formatted in JSON.

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

`<prefix>/audit/<YYYYMMDD>/<16-digit HEX string>`
------------------------------------------------

Each entry of audit log is stored with this type of key.
The value is JSON having fields defined in [audit log](audit.md).

* `<YYYYMMDD>` is the date of the event.
* `<16-digit HEX string>` is the hexadecimal representation of the event revision.

`<prefix>/audit`
----------------

This key stores RFC3339-format timestamp to record the last compaction
of audit logs.

`<prefix>/kernel-params/coreos`
----------------

This type of key holds kernel parameters.
