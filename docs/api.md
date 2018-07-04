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

```console
$ curl -XPUT 'localhost:10080/api/v1/config/ipam' -d '
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
- HTTP response header: `application/json`
- HTTP response body: JSON

**Failure responses**

- IPAM configurations have not been created

  HTTP status code: 404 Not Found

```console
$ curl -XGET 'localhost:10080/api/v1/config/ipam'
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

```console
$ curl -XPUT 'localhost:10080/api/v1/config/dhcp' -d '
{
    "gateway-offset": 1
}'
```

## <a name="getdhcp" />`GET /api/v1/config/dhcp`

Get DHCP configurations.

**Successful response**

- HTTP status code: 200 OK
- HTTP response header: `application/json`
- HTTP response body: JSON

**Failure responses**

- DHCP configuration have not been craeted

  HTTP status code: 404 Not Found

```console
$ curl -XGET 'localhost:10080/api/v1/config/dhcp'
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
`datacenter=<datacenter>`    | The data center name where the machine is in
`rack=<rack>`                | The rack number where the machine is in. If it is omitted, value set to `0`
`role=<role>`                | The role of the machine (e.g. `boot` or `worker`)
`product=<product>`          | The product name of the machine (e.g. `R630`)
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

```console
$ curl -X POST -H "Content-Type:application/json" 'localhost:10080/api/v1/machines' -d '
[{
  "serial": "1234abcd",
  "product": "R630",
  "datacenter": "ty3",
  "rack": 1,
  "role": "boot",
  "bmc": {"type": "iDRAC-9"},
}]'
```

## <a name="getmachines" />`GET /api/v1/machines`

Search registered machines. A user can specify the following URL queries.

Query                      | Description
-----                      | -----------
`serial=<serial>`          | The serial number of the machine
`datacenter=<datacenter>`  | The data center name where the machine is in
`rack=<rack>`              | The rack number where the machine is in
`role=<role>`              | The role of the machine
`product=<product>`        | The product name of the machine(e.g. `R630`)
`ipv4=<ip address>`        | IPv4 address
`ipv6=<ip address>`        | IPv6 address
`bmc-type=<bmc-type>`      | BMC type
`state=<state>`            | The state of the machine

**Successful response**

- HTTP status code: 200 OK
- HTTP response header: `application/json`
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

```console
$ curl -X DELETE 'localhost:10080/api/v1/machines/1234abcd'
(No output in stdout)
```

## <a name="putstate" />`PUT /api/v1/state/<serial>`

Put the state of a machine.
The new state is given by contents of request body and should be one of:
* healthy
* unhealthy
* dead
* retiring

**Successful response**

- HTTP status code: 200 OK

**Failure responses**

- Invalid state value.

  HTTP status code: 400 Bad Request

- No specified machine found.

  HTTP status code: 404 Not Found

## <a name="getstate" />`GET /api/v1/state/<serial>`

Get the state of a machine.
The state will be returned by response body and should be one of:
* healthy
* unhealthy
* dead
* retiring
* retired

**Successful response**
- HTTP status code: 200 OK

**Failure responses**
- No specified machine found.

  HTTP status code: 404 Not Found

## <a name="getimageindex" />`GET /api/v1/images/coreos`

Get the [image index](image_management.md).

## <a name="putimages" />`PUT /api/v1/images/coreos/<id>`

Upload a tar archive of CoreOS Container Linux boot image.
The tar file must consist of these two files:

* `kernel`: Linux kernel image.
* `initrd.gz`: Initial rootfs image.

**Successful response**

- HTTP status code: 201 Created
- HTTP response header: `application/json`
- HTTP response body: JSON

**Failure responses**

- An image having the same ID has already been registered in the index.

  HTTP status code: 409 Conflict

- Invalid tar image or invalid ID.

  HTTP status code: 400 Bad Request

## <a name="getimages" />`GET /api/v1/images/coreos/<id>`

Download the image archive specified by `<id>`.
The archive format is the same as PUT; i.e. a tar consists of `kernel` and `initrd.gz`.

**Successful response**

- HTTP status code: 200 OK
- HTTP response header: `application/tar`

**Failure responses**

- No image has the ID.

  HTTP status code: 404 Not found

## <a name="deleteimages" />`DELETE /api/v1/images/coreos/<id>`

Remove the image specified by `<id>` from the index.

**Successful response**

- HTTP status code: 200 OK
- HTTP response body: empty

**Failure responses**

- No image has the ID.

  HTTP status code: 404 Not found

## <a name="getassetsindex" />`GET /api/v1/assets`

Get the list of asset names as JSON array.

## <a name="putassets" />`PUT /api/v1/assets/<NAME>`

Upload a file as an asset.

**Request headers**

- `Content-Type`: required
- `Content-Length`: required
- `X-Sabakan-Asset-SHA256`: if given, the uploaded data will be verified by SHA256.

**Successful response**

- HTTP status code: 201 Created, or 200 OK
- HTTP response header: `application/json`
- HTTP response body: JSON

The response for a newly created asset looks like:
```json
{
    "status": 201,
    "id": "15"
}
```

The response for an updated asset looks like:
```json
{
    "status": 200,
    "id": "19"
}
```

**Failure responses**

- No content-type request header:

    HTTP status code: 400 Bad Request

- Upload conflicted:

    HTTP status code: 409 Conflicted

- No content-length request header:

    HTTP status code: 411 Length Required

- Content is too large:

    HTTP status code: 413 Payload Too Large

## <a name="getassets" />`GET /api/v1/assets/<NAME>`

Download the named asset.

**Successful response**

HTTP status code:
- 200 OK

Response headers:

- `X-Sabakan-Asset-ID`: ID of the asset
- `X-Sabakan-Asset-SHA256`: SHA256 checksum of the asset

**Failure responses**

- The asset was not found.

    HTTP status code: 404 Not found

## <a name="getassetsmeta" />`GET /api/v1/assets/<NAME>/meta`

Fetch the meta data of the named asset.

**Successful response**

HTTP status code:
- 200 OK

HTTP response header:
- `Content-Type`: `application/json`

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

```console
$ curl -XGET localhost:10080/api/v1/boot/ignitions/1234abcd/1527731687
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

```console
$ curl -XGET localhost:10080/api/v1/boot/ignitions/cs
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

```console
$ curl -XGET localhost:10080/api/v1/ignitions/cs/1527731687
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
- HTTP response header: `application/json`
- HTTP response body: JSON

```json
{"status": 201, "role": "<role>", "id": "<id>"}
```

**Failure responses**

- Invalid ignition format.

  HTTP status code: 400 Bad Request

- Invalid `<role>`.

  HTTP status code: 400 Bad Request

-
```console
$ curl -XPOST -d $'ignition:\n version: "2.2.0' localhost:10080/api/v1/ignitions/cs
{"status": 201, "role": "cs", "id": "1507731659"}
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

```console
$ curl -XDELETE localhost:10080/api/v1/boot/ignitions/cs/1527731687
```

## <a name="putcrypts" />`PUT /api/v1/crypts/<serial>/<path>`

Register disk encryption key. The request body is raw binary format of the key.

**Successful response**

- HTTP status code: 201 Created
- HTTP response header: `application/json`
- HTTP response body: JSON

```json
{"status": 201, "path": "<path>"}
```

**Failure responses**

- The machine is not found.

    HTTP status code: 404 Not Found

- The state of the machine is `retired`.

    HTTP status code: 500 Internal Server Error

- `/<prefix>/crypts/<serial>/<path>` already exists.

    HTTP status code: 409 Conflict

- The request body is empty.

  HTTP status code: 400 Bad Request

```console
$ head -c256 /dev/urandom | curl -i -X PUT -d - 'localhost:10080/api/v1/crypts/1/aaaaa'
HTTP/1.1 201 Created
Content-Type: application/json
Date: Tue, 10 Apr 2018 09:12:12 GMT
Content-Length: 31

{"status": 201, "path":"aaaaa"}
```

## <a name="getcrypts" />`GET /api/v1/crypts/<serial>/<path>`

Get an encryption key of the particular disk.

**Successful response**

- HTTP status code: 200 OK
- HTTP response header: `application/octet-stream`
- HTTP response body: A raw key data

**Failure responses**

- No specified `/<prefix>/crypts/<serial>/<path>` found in etcd.

  HTTP status code: 404 Not Found

```console
$ curl -i -X GET 'localhost:10080/api/v1/crypts/1/aaaaa'
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
- HTTP response header: `application/json`
- HTTP response body: Array of the `<path>` which are deleted successfully.

```console
$ curl -i -X DELETE 'localhost:10080/api/v1/crypts/1'
HTTP/1.1 200 OK
Content-Type: application/json
Date: Tue, 10 Apr 2018 09:19:01 GMT
Content-Length: 18

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

### Example:

`GET /api/v1/logs?since=20180404&until=20180407` retrieves logs generated on
2018-04-04, 2018-04-05, and 2018-04-06.  Note that the date specified for
`until` is not included.
