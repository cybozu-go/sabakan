sabactl
=======

Usage
-----

```console
$ sabactl [--server http://localhost:10080] <subcommand> <args>...
```

Option     | Default value            | Description
------     | -------------            | -----------
`--server` | `http://localhost:10080` | URL of sabakan

`sabactl ipam set`
------------------

Set/update IPAM configurations.  See [IPAMConfig](ipam.md#ipamconfig) for JSON fields.

```console
$ sabactl ipam set -f <ipam_configurations.json>
```

`sabactl ipam get`
------------------

Get IPAM configurations.

```console
$ sabactl ipam get
```

`sabactl dhcp set`
------------------

Set/update DHCP configurations.  See [DHCPConfig](dhcp.md#dhcpconfig) for JSON fields.

```console
$ sabactl dhcp set -f <dhcp_configurations.json>
```

`sabactl dhcp get`
------------------

Get DHCP configurations.

```console
$ sabactl dhcp get
```

`sabactl machines create`
-------------------------

Register new machines.

```console
$ sabactl machines create -f <machine_informations.json>
```

You can register multiple machines by giving a list of machine specs as shown below.
Detailed specification of the input JSON file is same as that of the [`POST /api/v1/machines` API](api.md#postmachines).

```json
[
  { "serial": "<serial1>", "datacenter": "<datacenter1>", "rack": "<rack1>", "product": "<product1>", "role": "<role1>", "bmc": { "type": "iDRAC-9" }},
  { "serial": "<serial2>", "datacenter": "<datacenter2>", "rack": "<rack2>", "product": "<product2>", "role": "<role2>", "bmc": { "type": "iDRAC-9" }}
]
```

`sabactl machines get`
----------------------

Show machines filtered by query parameters.

```console
$ sabactl machines get [--serial <serial>] [--state <state>] [--datacenter <datacenter>] [--rack <rack>] [--product <product>] [--ipv4 <ip address>] [--ipv6 <ip address>] [--bmc-type <BMC type>]
```

Detailed specification of the query parameters and the output JSON content is same as those of the [`GET /api/v1/machines` API](api.md#getmachines).

!!! Note
    `--state <state>` will not be implemented until the policy of machines life-cycle management is fixed.

`sabactl machines remove`
-------------------------

Unregister a machine specified by a serial number.

```console
$ sabactl machines remove <serial>
```

!!! Note
    This will be refined for machines life-cycle management.
    We should not unregister machines by their serials, but by their statuses.
    We can unregister machines only if their statuses are "to be repaired" or "to be discarded" or anythin like those.
    So the parameters of this command should be `--state <state>`.

`sabactl images [-os OS] index`
-------------------------------

Get the current index of the OS images as JSON.

* `-os`: specifies OS of the image.  Default is "coreos"

`sabactl images [-os OS] upload`
--------------------------------

```console
$ sabactl images upload ID coreos_production_pxe.vmlinuz coreos_production_pxe_image.cpio.gz
```

Upload a set of boot image files identified by `ID`.

* `-os`: specifies OS of the image.  Default is "coreos"

`sabactl images [-os OS] delete`
--------------------------------

```console
$ sabactl images delete ID
```

Delete an image.

* `-os`: specifies OS of the image.  Default is "coreos"

`sabactl assets index`
----------------------

Get the index of assets as a JSON array of asset names.

`sabactl assets info NAME`
--------------------------

Get the meta data of the named asset.

`sabactl assets upload NAME FILE`
---------------------------------

```console
$ sabactl assets upload data.tar.gz /path/to/data.tar.gz
```

Upload an asset.  NAME is the asset filename.
The data is read from FILE.

`sabactl assets delete NAME`
----------------------------

```console
$ sabactl assets delete data.tar.gz
```

Delete an asset.

`sabactl ignitions get ROLE`
----------------------------

Get a registered ignition template ID list of the role.

```console
$ sabactl ignitions get <role>
```

`sabactl ignitions cat ROLE ID`
-------------------------------

Get a registered ignition template of ID in the role. 

```console
$ sabactl ignitions cat <role> <id>
```

`sabactl ignitions set ROLE`
----------------------------

Register a new ignition template for a certain role.  The format ignitions are described in [Ignition Controls](ignition.md).

```console
$ sabactl ignitions set -f <ignition.yml> <role>
```

`sabactl ignitions delete ROLE ID`
----------------------------------

Delete a ignition template for a certain role.

```console
$ sabactl ignitions delete <role> <id>
```
