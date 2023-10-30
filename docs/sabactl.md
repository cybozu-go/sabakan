sabactl
=======

Usage
-----

```console
$ sabactl [--server http://localhost:10080] <subcommand> <args>...
```

| Option         | Default value            | Description                          |
| -------------- | ------------------------ | ------------------------------------ |
| `--server`     | `http://localhost:10080` | URL of sabakan server                |
| `--tls-server` | `https://localhost:10443`| URL of sabakan TLS server            |
| `--insecure`   | `false`                  | Disable TLS certificate verification |

`sabactl ipam set -f FILE`
--------------------------

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

`sabactl dhcp set -f FILE`
--------------------------

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

`sabactl machines create -f FILE`
---------------------------------

Register new machines.

```console
$ sabactl machines create -f <machine_informations.json>
```

You can register multiple machines by giving a list of machine specs as shown below.
Detailed specification of the input JSON file is same as that of the [`POST /api/v1/machines` API](api.md#postmachines).

```json
[
  { "serial": "<serial1>", "labels": {"product": "<product1>", "datacenter": "<datacenter1>"}, "rack": "<rack1>", "role": "<role1>", "bmc": { "type": "iDRAC-9" }},
  { "serial": "<serial2>", "labels": {"product": "<product2>", "datacenter": "<datacenter2>"}, "rack": "<rack2>", "role": "<role2>", "bmc": { "type": "iDRAC-9" }}
]
```

`sabactl machines get [QUERY_PARAM]...`
---------------------------------------

Show machines filtered by query parameters.

```console
$ sabactl machines get \
    [--serial <serial>,...] \
    [--rack <rack>,...] \
    [--role <role>,...] \
    [--labels <key=value>,...]
    [--ipv4 <ip address>,...] \
    [--ipv6 <ip address>,...] \
    [--bmc-type <BMC type>,...] \
    [--state <state>,...] \
    [--without-serial <serial>,...] \
    [--without-rack <rack>,...] \
    [--without-role <role>,...] \
    [--without-labels <key=value>,...]
    [--without-ipv4 <ip address>,...] \
    [--without-ipv6 <ip address>,...] \
    [--without-bmc-type <BMC type>,...] \
    [--without-state <state>,...] \
    [--output json|simple]
```

Detailed specification of the query parameters and the output JSON content is same as those of the [`GET /api/v1/machines` API](api.md#getmachines).

`sabactl machines set-label SERIAL NAME VALUE`
----------------------------------------------

Add or update a label for a machine.

```console
$ sabactl machines set-label <serial> <name> <value>
```

`sabactl machines remove-label SERIAL NAME`
-------------------------------------------

Remove a label from a machine.

```console
$ sabactl machines remove-label <serial> <name>
```

`sabactl machines set-retire-date SERIAL DATE`
----------------------------------------------

Set the retire date of a machine.

```console
$ sabactl machines set-retire-date <serial> 2023-11-21
```

`sabactl machines set-state SERIAL STATE`
-----------------------------------------

Set the state of a machine.
State is one of `uninitialized`, `healthy`, `unhealthy`, `unreachable`, `updating`, `retiring` or `retired`.

Transition from `retiring` to `retired` is permitted only when the machine has no disk encryption keys.

```console
$ sabactl machines set-state <serial> <state>
```

`sabactl machines get-state SERIAL`
-----------------------------------

Get the state of a machine.
State is one of `uninitialized`, `healthy`, `unhealthy`, `unreachable`, `updating`, `retiring` or `retired`.

```console
$ sabactl machines get-state <serial>
```

`sabactl machines remove SERIAL`
--------------------------------

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

* `--os`: specifies OS of the image.  Default is "coreos"

`sabactl images [-os OS] upload ID KERNEL INITRD`
-------------------------------------------------

```console
$ sabactl images upload ID coreos_production_pxe.vmlinuz coreos_production_pxe_image.cpio.gz
```

Upload a set of boot image files identified by `ID`.

* `--os`: specifies OS of the image.  Default is "coreos"

!!! Note
    You can execute upload multiple times for a certain ID only with the same set of files.

!!! Note
    Once the set of boot image files is deleted, no matter if manually or automatically, you cannot upload with the same ID.

`sabactl images [-os OS] delete ID`
------------------------------------

```console
$ sabactl images delete ID
```

Delete an image.

* `--os`: specifies OS of the image.  Default is "coreos"

`sabactl assets index`
----------------------

Get the index of assets as a JSON array of asset names.

`sabactl assets info NAME`
--------------------------

Get the meta data of the named asset.

`sabactl assets upload [--meta KEY=VALUE]... NAME FILE`
-------------------------------------------------------

```console
$ sabactl assets upload data.tar.gz /path/to/data.tar.gz
```

Upload an asset.  NAME is the asset filename.
The data is read from FILE.

* `--meta`: adds meta data.

`sabactl assets delete NAME`
----------------------------

```console
$ sabactl assets delete data.tar.gz
```

Delete an asset.

`sabactl ignitions get ROLE [ID]`
---------------------------------

List or get ignition templates for `ROLE`.

If `ID` is not specified, the command outputs the list of available `ID` for `ROLE`.

If `ID` is specified, the command outputs a JSON object like this:
```json
{
  "version": "2.3",
  "template": "{\"ignition\": {\"version\": \"2.3.0\"}, ...}",
  "meta": "{\"foo\": [1, 2, 3]}"
}
```

`sabactl ignitions set -f FILE [--json] [--meta FILENAME] ROLE ID`
------------------------------------------------------------------

Register a new ignition template for `ROLE` with `ID`.
For details, see .

`FILE` is either a template YAML described in [Ignition Templates](ignition_template.md), or
a JSON got by `sabactl ignitions get ROLE ID` if `--json` is given.

A template can be associated with meta data if `--meta FILENAME` is given.
The meta data file must contain a JSON object.  Meta data can be referenced
when rendering the ignition template.

```console
$ sabactl ignitions set -f compute.yml <role> <id>
```

`sabactl ignitions delete ROLE ID`
----------------------------------

Delete a ignition template for a certain role.

```console
$ sabactl ignitions delete <role> <id>
```

`sabactl log [--json] [START_DATE] [END_DATE]`
----------------------------------------------

Retrieve [audit logs](audit.md) and output them to stdout.

If `--json` is given, each log entry will be displayed in JSON.

If `START_DATE` is given, and `END_DATE` is *not* given, logs
of `START_DATE` are retrieved.

If `START_DATE` and `END_DATE` is given, logs between them are
retrieved.

`sabactl kernel-params [-os OS] set PARAMS`
-------------------------------------------

Set/update kernel parameters.

* `--os`: specifies OS of the image.  Default is "coreos"

```console
$ sabactl kernel-params set "<param0>=<value0> <param1>=<value1> ..."
```

`sabactl kernel-params [-os OS] get`
------------------------------------

Get the current kernel parameters.

* `--os`: specifies OS of the image.  Default is "coreos"

```console
$ sabactl kernel-params get
```

`sabactl crypts delete SERIAL`
------------------------------

Deletes all keys of a machine.
The command fails when the target machine's status is not `retiring`.

- `--force`: explicitly required

```console
$ sabactl crypts delete -force <serial>
```

`sabactl version`
-----------------

Show sabactl version

```console
$ sabactl version
```
