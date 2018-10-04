REST API
========

* [PUT /api/v1/config/ipam](#putipam)
* [GET /api/v1/config/ipam](#getipam)
* [PUT /api/v1/config/dhcp](#putdhcp)
* [GET /api/v1/config/dhcp](#getdhcp)
* [POST /api/v1/machines](#postmachines)
* [GET /api/v1/machines](#getmachines)
* [DELETE /api/v1/machines](#deletemachines)
* [PUT /api/v1/state/\<serial\>](#putstate)
* [GET /api/v1/state/\<serial\>](#getstate)
* [PUT /api/v1/labels/\<serial\>](#putlabels)
* [DELETE /api/v1/labels/\<serial\>](#deletelabels)
* [GET /api/v1/images/coreos](#getimageindex)
* [PUT /api/v1/images/coreos/\<id\>](#putimages)
* [GET /api/v1/images/coreos/\<id\>](#getimages)
* [DELETE /api/v1/images/coreos/\<id\>](#deleteimages)
* [GET /api/v1/assets](#getassetsindex)
* [PUT /api/v1/assets/\<name\>](#putassets)
* [GET|HEAD /api/v1/assets/\<name\>](#getassets)
* [GET /api/v1/assets/\<name\>/meta](#getassetsmeta)
* [DELETE /api/v1/assets/\<name\>](#deleteassets)
* [GET /api/v1/boot/ipxe.efi](#getipxe)
* [GET /api/v1/boot/coreos/ipxe](#getcoreosipxe)
* [GET /api/v1/boot/coreos/ipxe/\<serial\>](#getcoreosipxeserial)
* [GET|HEAD /api/v1/boot/coreos/kernel](#getcoreoskernel)
* [GET|HEAD /api/v1/boot/coreos/initrd.gz](#getcoreosinitrd)
* [GET /api/v1/boot/ignitions/\<serial\>/\<id\>](#getigitionsid)
* [GET /api/v1/ignitions/\<role\>](#getignitions)
* [GET /api/v1/ignitions/\<role\>/\<id\>](#getignitionsid)
* [POST /api/v1/ignitions/\<role\>](#postignitions)
* [DELETE /api/v1/ignitions/\<role\>/\<id\>](#deleteignitions)
* [PUT /api/v1/crypts](#putcrypts)
* [GET /api/v1/crypts](#getcrypts)
* [DELETE /api/v1/crypts](#deletecrypts)
* [GET /api/v1/logs](#getlogs)
* [PUT /api/v1/kernel_params/coreos](#putkernelparams)
* [GET /api/v1/kernel_params/coreos](#getkernelparams)
* [GET /version](#version)
* [GET /health](#health)

## Access control

The following requets URLs are allowed for all remote hosts.  The other URLs
are rejected from remote hosts excluding addresses specified in `-allow-ips` option.

- `PUT /api/v1/crypts`
- `GET /api/v1/crypts`
- `GET|HEAD /*`

This means that localhost can manage all resources, and the remote hosts such
as worker nodes can only read resources.  `PUT /api/v1/crypts` and `GET
/api/v1/crypts` are permitted from all remote hosts since the encryption keys
are generated on the client nodes.  The encryption keys *should* be distributed
between sabakan nodes and the client node.

## <a name="putipam" />`PUT /api/v1/config/ipam`

Create or update IPAM configurations.  If one or more nodes have been registered in sabakan, IPAM configurations cannot be updated.

**Successful response**

- HTTP status code: 200 OK

**Failure responses**

- One or more nodes are already registered.

  HTTP status code: 500 Internal Server Error

**Example**

```console
$ curl -s -XPUT 'localhost:10080/api/v1/config/ipam' -d '
{
   "max-nodes-in-rack": 28,
   "node-ipv4-pool": "10.69.0.0/16",
   "node-ipv4-range-size": 6,
   "node-ipv4-range-mask": 26,
   "node-ip-per-node": 3,
   "node-index-offset": 3,
   "bmc-ipv4-pool": "10.72.16.0/20",
   "bmc-ipv4-offset": "0.0.1.0",
   "bmc-ipv4-range-size": 5,
   "bmc-ipv4-range-mask": 20
}'
```

## <a name="getipam" />`GET /api/v1/config/ipam`

Get IPAM configurations.

The body must be JSON representation of [IPAMConfig](ipam.md#ipamconfig).

**Successful response**

- HTTP status code: 200 OK
- HTTP response header: `Content-Type: application/json`
- HTTP response body: Current IPAM configurations in JSON

**Failure responses**

- IPAM configurations have not been created

  HTTP status code: 404 Not Found

**Example**

```console
$ curl -s -XGET 'localhost:10080/api/v1/config/ipam'
{
   "max-nodes-in-rack": 28,
   "node-ipv4-pool": "10.69.0.0/16",
   "node-ipv4-range-size": 6,
   "node-ipv4-range-mask": 26,
   "node-ip-per-node": 3,
   "node-index-offset": 3,
   "bmc-ipv4-pool": "10.72.16.0/20",
   "bmc-ipv4-offset": "0.0.1.0",
   "bmc-ipv4-range-size": 5,
   "bmc-ipv4-range-mask": 20
}
```

## <a name="putdhcp" />`PUT /api/v1/config/dhcp`

Create or update DHCP configurations.

The body must be JSON representation of [DHCPConfig](dhcp.md#dhcpconfig).

**Successful response**

- HTTP status code: 200 OK

**Failure responses**

- HTTP status codes other than 200.

**Example**

```console
$ curl -s -XPUT 'localhost:10080/api/v1/config/dhcp' -d '
{
    "gateway-offset": 1
}'
```

## <a name="getdhcp" />`GET /api/v1/config/dhcp`

Get DHCP configurations.

**Successful response**

- HTTP status code: 200 OK
- HTTP response header: `Content-Type: application/json`
- HTTP response body: Current DHCP configurations in JSON

**Failure responses**

- DHCP configuration have not been craeted

  HTTP status code: 404 Not Found

**Example**

```console
$ curl -s -XGET 'localhost:10080/api/v1/config/dhcp'
{
    "gateway-offset": 254
}
```

## <a name="postmachines" />`POST /api/v1/machines`

Register machines.
All of the machines in the requested JSON is an atomic operation to register.
If Sabakan fails to register at least one machine, it all fails. In other words, the result will be registered all machines or not registered at all.
There is no possibility that part of machines will be registered.

In the HTTP request body, specify the following list of the machine information in JSON format.

Field                        | Description
-----                        | -----------
`serial=<serial>`            | The serial number of the machine
`labels={<key=value>, ...}`    | The labels of the machine
`rack=<rack>`                | The rack number where the machine is in. If it is omitted, value set to `0`
`role=<role>`                | The role of the machine (e.g. `boot` or `worker`)
`bmc=<bmc>`                  | The BMC spec

**Successful response**

- HTTP status code: 201 Created

**Failure responses**

- The same serial number of the machine is already registered.

  HTTP status code: 409 Conflict

- The boot server in the specified `rack` is already registered.

  HTTP status code: 409 Conflict

- Invalid value of `<role>` format.

  HTTP status code: 400 Bad Request

**Example**

```console
$ curl -s -X POST 'localhost:10080/api/v1/machines' -d '
[
  { "serial": "1234abcd", "labels": {"product": "R630", "datacenter": "ty3"}, "rack": 1, "role": "boot", "bmc": {"type": "iDRAC-9"} },
  { "serial": "2345bcde", "labels": {"product": "R630", "datacenter": "ty3"}, "rack": 1, "role": "worker", "bmc": {"type": "iDRAC-9"} }
]'
```

## <a name="getmachines" />`GET /api/v1/machines`

Search registered machines. A user can specify the following URL queries.

Query                      | Description
-----                      | -----------
`serial=<serial>`          | The serial number of the machine
`labels=<key=value>,...`   | The labels of the machine.
`rack=<rack>`              | The rack number where the machine is in
`role=<role>`              | The role of the machine
`ipv4=<ip address>`        | IPv4 address
`ipv6=<ip address>`        | IPv6 address
`bmc-type=<bmc-type>`      | BMC type
`state=<state>`            | The state of the machine

**Successful response**

- HTTP status code: 200 OK
- HTTP response header: `Content-Type: application/json`
- HTTP response body: Machines of an array of the JSON

**Failure responses**

- No such machines found.

  HTTP status code: 404 Not Found

## <a name="deletemachines" />`DELETE /api/v1/machines/<serial>`

Delete registered machine of the `<serial>`.

**Successful response**

- HTTP status code: 200 OK
- HTTP response body: empty

**Failure responses**

- The state of the machine is not `retired`

  HTTP status code: 500 Internal Server Error

- No specified machine found.

  HTTP status code: 404 Not Found

**Example**

```console
$ curl -s -X DELETE 'localhost:10080/api/v1/machines/1234abcd'
(No output in stdout)
```

## <a name="putstate" />`PUT /api/v1/state/<serial>`

Put the state of a machine.
The new state is given by contents of request body and should be one of:

* `uninitialized`
* `healthy`
* `unhealthy`
* `dead`
* `updating`
* `retiring`

**Successful response**

- HTTP status code: 200 OK

**Failure responses**

- Invalid state value.

  HTTP status code: 400 Bad Request

- No specified machine found.

  HTTP status code: 404 Not Found

- Invalid state transition

  HTTP status code: 500 Internal Server Error

**Example**

```console
$ curl -s -XPUT -d'retiring' localhost:10080/api/v1/state/1234abcd
(No output in stdout)
```

## <a name="getstate" />`GET /api/v1/state/<serial>`

Get the state of a machine.
The state will be returned by response body and should be one of:
* uninitialized
* healthy
* unhealthy
* dead
* updating
* retiring
* retired

**Successful response**
- HTTP status code: 200 OK

**Failure responses**
- No specified machine found.

  HTTP status code: 404 Not Found

**Example**

```console
$ curl -s localhost:10080/api/v1/state/1234abcd
retiring
```

## <a name="putlabels" />`PUT /api/v1/labels/<serial>`

Add labels to a machine. A value is overwritten when the label already exists.

**Successful response**

- HTTP status code: 200 OK

**Failure responses**

- Invalid label format.

  HTTP status code: 400 Bad Request

- No specified machine found.

  HTTP status code: 404 Not Found

**Example**

```console
$ curl -s -XPUT localhost:10080/api/v1/labels/1234abcd -d '
{
    "os-release": "1855.4.0"
}
'
(No output in stdout)
```

## <a name="deletelabels" />`DELETE /api/v1/labels/<serial>/<label>`

Remove label from a machine.

**Successful response**

- HTTP status code: 200 OK
- HTTP response body: empty

**Failure responses**

- No label has in the machine.

  HTTP status code: 404 Not found

**Example**

```console
$ curl -s -XDELETE 'localhost:10080/api/v1/labels/1234abcd/os-release'
(No output in stdout)
```

## <a name="getassetsindex" />`GET /api/v1/assets`

Get the list of asset names as JSON array.

**Successful response**

- HTTP status code: 200 OK
- HTTP response header: `Content-Type: application/json`
- HTTP response body: Index of the registered assets in JSON

**Example**

```console
$ curl -s 'localhost:10080/api/v1/assets'
[
  "cybozu-bird-2.0.aci",
  "cybozu-chrony-3.3.aci",
  "cybozu-ubuntu-debug-18.04.aci",
  "sabakan-cryptsetup"
]
```

## <a name="getimageindex" />`GET /api/v1/images/coreos`

Get the [image index](image_management.md) for coreos.

**Successful response**

- HTTP status code: 200 OK
- HTTP response header: `Content-Type: application/json`
- HTTP response body: Index of the registered images in JSON

**Example**

```console
$ curl -s localhost:10080/api/v1/images/coreos
[
  {
    "id": "1745.5.0",
    "date": "2018-07-04T23:26:01.392249742Z",
    "urls": [
      "http://10.69.0.3:10080/api/v1/images/coreos/1745.5.0",
      "http://10.69.0.195:10080/api/v1/images/coreos/1745.5.0",
      "http://10.69.1.131:10080/api/v1/images/coreos/1745.5.0"
    ],
    "exists": true
  }
]
```


## <a name="putimages" />`PUT /api/v1/images/coreos/<id>`

Upload a tar archive of CoreOS Container Linux boot image.
The tar file must consist of these two files:

* `kernel`: Linux kernel image.
* `initrd.gz`: Initial rootfs image.

**Successful response**

- HTTP status code: 201 Created
- HTTP response body: empty

**Failure responses**

- An image having the same ID has already been registered in the index.

  HTTP status code: 409 Conflict

- Invalid tar image or invalid ID.

  HTTP status code: 400 Bad Request

**Example**

```console
$ curl -s -XPUT --data-binary '@./path/to/coreos-image.tar' 'localhost:10080/api/v1/images/coreos/1745.7.0'
(No output in stdout)
```

## <a name="getimages" />`GET /api/v1/images/coreos/<id>`

Download the image archive specified by `<id>`.
The archive format is the same as PUT; i.e. a tar consists of `kernel` and `initrd.gz`.

**Successful response**

- HTTP status code: 200 OK
- HTTP response header: `Content-Type: application/tar`
- HTTP response body: image archive in tar binary

**Failure responses**

- No image has the ID.

  HTTP status code: 404 Not found

```console
$ curl -s -i 'localhost:10080/api/v1/images/coreos/1745.7.0'
HTTP/1.1 200 OK
Content-Type: application/tar
Date: Wed, 04 Jul 2018 23:58:36 GMT
Transfer-Encoding: chunked

.....
```

## <a name="deleteimages" />`DELETE /api/v1/images/coreos/<id>`

Remove the image specified by `<id>` from the index.

**Successful response**

- HTTP status code: 200 OK
- HTTP response body: empty

**Failure responses**

- No image has the ID.

  HTTP status code: 404 Not found

**Example**

```console
$ curl -s -XDELETE 'localhost:10080/api/v1/images/coreos/1688.5.3'
(No output in stdout)
```

## <a name="getassetsindex" />`GET /api/v1/assets`

Get the list of asset names as JSON array.

**Successful response**

- HTTP status code: 200 OK
- HTTP response header: `Content-Type: application/json`
- HTTP response body: Index of the registered assets in JSON

**Example**

```console
$ curl -s 'localhost:10080/api/v1/assets'
[
  "cybozu-bird-2.0.aci",
  "cybozu-chrony-3.3.aci",
  "cybozu-ubuntu-debug-18.04.aci",
  "sabakan-cryptsetup"
]
```

## <a name="putassets" />`PUT /api/v1/assets/<NAME>`

Upload a file as an asset.

**Request headers**

- `Content-Type`: required
- `Content-Length`: required
- `X-Sabakan-Asset-SHA256`: if given, the uploaded data will be verified by SHA256.

**Successful response**

- HTTP status code: 201 Created, or 200 OK
- HTTP response header: `Content-Type: application/json`
- HTTP response body: Asset's ID in JSON

**Failure responses**

- No content-type request header:

    HTTP status code: 400 Bad Request

- Upload conflicted:

    HTTP status code: 409 Conflicted

- No content-length request header:

    HTTP status code: 411 Length Required

- Content is too large:

    HTTP status code: 413 Payload Too Large

**Example**

The response for a newly created asset looks like:

```console
$ curl -s -XPUT --data-binary '@./sabakan-cryuptsetup' 'localhost:10080/api/v1/assets/sabakan-cryptsetup'
{
    "status": 201,
    "id": "15"
}
```

The response for an updated asset looks like:

```console
$ curl -s -XPUT --data-binary '@./sabakan-cryuptsetup' 'localhost:10080/api/v1/assets/sabakan-cryptsetup'
{
    "status": 200,
    "id": "19"
}
```

## <a name="getassets" />`GET /api/v1/assets/<NAME>`

Download the named asset.

**Successful response**

- HTTP status code: 200 OK
- HTTP Response headers:
    - `X-Sabakan-Asset-ID`: ID of the asset
    - `X-Sabakan-Asset-SHA256`: SHA256 checksum of the asset

**Failure responses**

- The asset was not found.

    HTTP status code: 404 Not found

## <a name="getassetsmeta" />`GET /api/v1/assets/<NAME>/meta`

Fetch the meta data of the named asset.

**Successful response**

- HTTP status code: 200 OK
- HTTP response header: `Content-Type: application/json`

The response JSON is described in [asset management](assets.md).

**Failure responses**

- The asset was not found.

    HTTP status code: 404 Not found

## <a name="deleteassets" />`DELETE /api/v1/assets/<NAME>`

Remove the named asset.

**Successful response**

- HTTP status code: 200 OK
- HTTP response body: empty

**Failure responses**

- The asset was not found.

    HTTP status code: 404 Not found

## <a name="getipxe" />`GET /api/v1/boot/ipxe.efi`

Get `ipxe.efi` firmware.

## <a name="getcoreosipxe" />`GET /api/v1/boot/coreos/ipxe`

Get iPXE script to chain URL to redirect  `/api/v1/boot/coreos/ipxe/<serial>`

## <a name="getcoreosipxeserial" />`GET /api/v1/boot/coreos/ipxe/<serial>`

Get iPXE script to boot CoreOS Container Linux.

Following query parameters can be added.

Name   | Value  | Description
----   | -----  | -----------
serial | 0 or 1 | serial console is enabled if 1

## <a name="getcoreoskernel" />`GET|HEAD /api/v1/boot/coreos/kernel`

Get Linux kernel image to boot CoreOS.

## <a name="getcoreosinitrd" />`GET|HEAD /api/v1/boot/coreos/initrd.gz`

Get initial RAM disk image to boot CoreOS.

## <a name="getigitionsid" />`GET /api/v1/boot/ignitions/<serial>/<id>`

Get CoreOS ignition for a certain serial.

**Successful response**

- HTTP status code: 200 OK

**Failure responses**

- No `<serial>` is found.

  HTTP status code: 404 Not found

- No ignition for `<id>` is found.

  HTTP status code: 404 Not found

**Example**

```console
$ curl -s -XGET localhost:10080/api/v1/boot/ignitions/1234abcd/1527731687
{
  "systemd": [
    ......
  ]
}
```

## <a name="getignitions" />`GET /api/v1/ignitions/<role>`

Get CoreOS ignition ids for a certain role.

**Successful response**

- HTTP status code: 200 OK

**Failure responses**

- No ignitions are registered in the role.

  HTTP status code: 404 Not found

- Invalid `<role>`.

  HTTP status code: 400 Bad Request

**Example**

```console
$ curl -s -XGET localhost:10080/api/v1/boot/ignitions/worker
[ "1427731487", "1507731659", "1527731687"]
```

## <a name="getignitionsid" />`GET /api/v1/ignitions/<role>/<id>`

Get CoreOS ignition template for a certain role.

**Successful response**

- HTTP status code: 200 OK

**Failure responses**

- No `<id>` exists in `<role>`.

  HTTP status code: 404 Not found

- Invalid `<role>`.

  HTTP status code: 400 Bad Request

**Example**

```console
$ curl -s -XGET localhost:10080/api/v1/ignitions/worker/1527731687
{
  "systemd": [
    ......
  ]
}
```

## <a name="postignitions" />`POST /api/v1/ignitions/<role>`

Create CoreOS ignition for a certain role from ignition-like YAML format (see [Ignition Controls](ignition.md)).
It returns a new assigned ID for the ignition.

**Successful response**

- HTTP status code: 201 Created
- HTTP response header: `Content-Type: application/json`
- HTTP response body: The target role of the ignition and ignition's ID in JSON

**Failure responses**

- Invalid ignition format.

  HTTP status code: 400 Bad Request

- Invalid `<role>`.

  HTTP status code: 400 Bad Request

**Example**

```console
$ echo $'ignition:\n version: "2.2.0"' | curl -s -XPOST -d - localhost:10080/api/v1/ignitions/worker
{"status": 201, "role": "worker", "id": "1507731659"}
```

## <a name="deleteignitions" />`DELETE /api/v1/ignitions/<role>/<id>`

Delete CoreOS ignition by role and id.

**Successful response**

- HTTP status code: 200 OK
- HTTP response body: empty

**Failure responses**

- Missing role or id

  HTTP status code: 400 Bad Request

- No `<id>` exists in `<role>`.

  HTTP status code: 404 Not found

- Invalid `<role>`.

  HTTP status code: 400 Bad Request

**Example**

```console
$ curl -s -XDELETE localhost:10080/api/v1/boot/ignitions/worker/1527731687
```

## <a name="putcrypts" />`PUT /api/v1/crypts/<serial>/<path>`

Register disk encryption key. The request body is raw binary format of the key.

**Successful response**

- HTTP status code: 201 Created
- HTTP response header: `Content-Type: application/json`
- HTTP response body: Registered path of the disk in JSON.

**Failure responses**

- The machine is not found.

    HTTP status code: 404 Not Found

- The state of the machine is `retired`.

    HTTP status code: 500 Internal Server Error

- `/<prefix>/crypts/<serial>/<path>` already exists in etcd.

    HTTP status code: 409 Conflict

- The request body is empty.

  HTTP status code: 400 Bad Request

**Example**

```console
$ head -c256 /dev/urandom | curl -s -i -X PUT -d - 'localhost:10080/api/v1/crypts/1/pci-0000:00:17.0-ata-1'
HTTP/1.1 201 Created
Content-Type: application/json
Date: Tue, 10 Apr 2018 09:12:12 GMT
Content-Length: 31

{"status": 201, "path":"pci-0000:00:17.0-ata-1"}
```

## <a name="getcrypts" />`GET /api/v1/crypts/<serial>/<path>`

Get an encryption key of the particular disk.

**Successful response**

- HTTP status code: 200 OK
- HTTP response header: `Content-Type: application/octet-stream`
- HTTP response body: A raw key data

**Failure responses**

- No specified `/<prefix>/crypts/<serial>/<path>` found in etcd in etcd.

  HTTP status code: 404 Not Found

**Example**

```console
$ curl -s -i 'localhost:10080/api/v1/crypts/1/pci-0000:00:17.0-ata-1'
HTTP/1.1 200 OK
Content-Type: application/octet-stream
Date: Tue, 10 Apr 2018 09:15:59 GMT
Content-Length: 64

.....
```

## <a name="deletecrypts" />`DELETE /api/v1/crypts/<serial>`

Delete all disk encryption keys of the specified machine. This request does not delete `/api/v1/machines/<serial>`, User can re-register encryption keys using `<serial>`.

**Successful response**

- HTTP status code: 200 OK
- HTTP response header: `Content-Type: application/json`
- HTTP response body: Array of the `<path>` which are deleted successfully in JSON.

**Example**

```console
$ curl -s -X DELETE 'localhost:10080/api/v1/crypts/1'
["abdef", "aaaaa"]
```

**Failure responses**

- The machine's state is not `retiring`.

    HTTP status code: 500 Internal Server Error

- The machine is not found.

    HTTP status code: 404 Not Found

## <a name="getlogs" />`GET /api/v1/logs`

Retrieve logs as [JSONLines](http://jsonlines.org/).
Each line represents an audit log entry as described in [audit log](audit.md).

If no URL parameter is given, this returns all logs stored in etcd.
Following parameters can be specified to limit the response:

* `since=YYYYMMDD`: retrieve logs after `YYYYMMDD`.
* `until=YYYYMMDD`: retrieve logs before `YYYYMMDD`.

The dates are interpreted in UTC timezone.

For example, `GET /api/v1/logs?since=20180404&until=20180407` retrieves logs
generated on 2018-04-04, 2018-04-05, and 2018-04-06.  Note that the date
specified for `until` is not included.

**Successful response**

- HTTP status code: 200 OK
- HTTP response header: `Content-Type: application/json`
- HTTP response body: Audit logs in JSONLines

**Example**

```console
$ curl -s 'localhost:10080/api/v1/logs?since=20180404&until=20180707'
{"ts":"2018-07-05T00:21:02.486519609Z","rev":"2","user":"root","ip":"127.0.0.1","host":"rkt-9ac1f230-64c4-40b1-a67c-52c297435d8b","category":"ipam","instance":"config","action":"put","detail":"{\"max-nodes-in-rack\":28,\"node-ipv4-pool\":\"10.69.0.0/20\",\"node-ipv4-range-size\":6,\"node-ipv4-range-mask\":26,\"node-ip-per-node\":3,\"node-index-offset\":3,\"bmc-ipv4-pool\":\"10.72.16.0/20\",\"bmc-ipv4-offset\":\"0.0.1.0\",\"bmc-ipv4-range-size\":5,\"bmc-ipv4-range-mask\":20}"}
{"ts":"2018-07-05T00:21:02.56064736Z","rev":"4","user":"root","ip":"127.0.0.1","host":"rkt-9ac1f230-64c4-40b1-a67c-52c297435d8b","category":"dhcp","instance":"config","action":"put","detail":"{\"gateway-offset\":1,\"lease-minutes\":60}"}
......
```

## <a name="putkernelparams" />`PUT /api/v1/kernel_params/coreos`

Create or update kernel parameters on iPXE booting.

**Successful response**

- HTTP status code: 200 OK

**Failure responses**

- Kernel params string contains non ASCII character(s) or control sequence(s).

  HTTP status code: 400 Bad Request

**Example**

```console
$ curl -s -XPUT 'localhost:10080/api/v1/kernel_params/coreos' -d '
console=ttyS0 coreos.autologin=ttyS0
'
```

## <a name="getkernelparams" />`GET /api/v1/kernel_params/coreos`

Get kernel parameters.

**Successful response**

- HTTP status code: 200 OK
- HTTP response header: `Content-Type: text/plain`
- HTTP response body: Current kernel parameters

**Failure responses**

- Kernel parameters have not been created

  HTTP status code: 404 Not Found

**Example**

```console
$ curl -s -XGET 'localhost:10080/api/v1/kernel_params/coreos'
console=ttyS0 coreos.autologin=ttyS0
```

## <a name="version" />`GET /version`

show sabakan version

**Successful response**

- HTTP status code: 200 OK
- HTTP response header: `Content-Type: application/json`
- HTTP response body: sabakan version

**Example**

```console
$ curl -s -XGET localhost:10080/version
{"version":"0.18"}
```

## <a name="health" />`GET /health`

get sabakan health status

**Successful response**
- HTTP status code: 200 OK
- HTTP response header: `Content-Type: application/json`
- HTTP response body: health status of sabakan

**Failure response**
- HTTP status code: 500 Internal Server Error
- HTTP response header: `application/json`
- HTTP response body: health status of sabakan

**Example**

```console
$ curl -s -XGET localhost:10080/health
{"health":"healthy"}
```

