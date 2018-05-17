Specifications
==============

Concepts
--------

### Node / Machine

The primary purpose of sabakan is to manage tons of physical servers in a data center.
A physical server is called a *node* or a *machine* in sabakan.

### BMC

Sabakan assumes each physical server equips a [baseboard management controller][BMC].
BMC is a dedicated hardware to manage the server remotely via network.

### Rack

Sabakan assumes that nodes are located in *racks*.  A rack can have limited number
of servers.  Sabakan divides a range of network address space into small ranges.
Each divided range is for a rack and should have enough capacity to allocate IP
address to all servers in the rack.

### Role

Each node has *role* attribute.  Roles can be defined arbitrary as string values,
except for **"boot"**.  Role "boot" is special; only one **boot** node can exist
in a rack.

### UEFI HTTP Boot

[UEFI HTTP Boot][HTTPBoot] is a modern network boot technology that replaces
legacy [PXE boot](https://en.wikipedia.org/wiki/Preboot_Execution_Environment).

UEFI HTTP Boot requires DHCP to configure the initial networking.  However,
unlike PXE, HTTP boot does not use TFTP to load boot loaders.  Instead of TFTP,
HTTP is used.

Sabakan is optimized for UEFI HTTP Boot.  It speaks DHCP and HTTP, but does
not speak TFTP.

### Ignition

[Ignition][] is a provisioning utility for [CoreOS Container Linux][CoreOS].
Sabakan can generate ignition config file for each node to provision CoreOS
Container Linux into nodes via network.

IP address management (IPAM)
----------------------------

### Overview

Sabakan assigns IP addresses to nodes for two purposes.  One is for DHCP, and another is for static assignment.  DHCP is used for the initial network boot.

Each machine may have one or more IP addresses for OS, and one for its [baseboard management controller][BMC].

### IPAMConfig

`IPAMConfig` is a set of configurations to assign IP addresses automatically.
It is given as JSON object with the following fields:

Field                  | Type   | Description
---------------------- | ------ | -----------
`max-nodes-in-rack`    | int    | The maximum number of nodes in a rack, excluding "boot" node.
`node-ipv4-pool`       | string | CIDR IPv4 network for node IP pool.
`node-ipv4-range-size` | int    | Size of the address range to divide the pool (bit counts).
`node-ipv4-range-mask` | int    | The subnet mask for a divided range.
`node-ip-per-node`     | int    | The number of IP addresses for each node.
`node-index-offset`    | int    | Offset for assigning IP address to a node in a divided range.
`bmc-ipv4-pool`        | string | CIDR IPv4 network for BMC IP pool.
`bmc-ipv4-range-size`  | int    | Size of the address range to divide the pool (bit counts).
`bmc-ipv4-range-mask`  | int    | The subnet mask for a divided range.

### Assigning static IPv4 addresses to Node OS

Sabakan computes static IP addresses for a node OS as follows (pseudo code):

```go
rack := node.RackNumber
idx  := node.IndexInRack

range_size := 1 << node-ipv4-range-size
base := INET_ATON(node-ipv4-pool)
addr0 := base + range_size * node-ip-per-node * rack + idx

addresses := []net.IP
for i := 0; i < node-ip-per-node; i++ {
    addresses = append(addresses, INET_NTOA(addr0 + i*range_size))
}
```

### DHCP lease range for Node OS

TODO

### Assigning static IPv4 address to BMC

Sabakan computes static IP addresses for BMC as follows (pseudo code):

```go
rack := node.RackNumber
idx  := node.IndexInRack

range_size := 1 << bmc-ipv4-range-size
base := INET_ATON(bmc-ipv4-pool)

bmc_addr := INET_NTOA(base + range_size * rack + idx)
```

### Examples

Suppose that IPAM configurations are as follows:

Field                  | Value
---------------------- | -----:
`max-nodes-in-rack`    | 28
`node-ipv4-pool`       | 10.69.0.0/16
`node-ipv4-range-size` | 6
`node-ipv4-range-mask` | 26
`node-ip-per-node`     | 3
`node-index-offset`    | 3
`bmc-ipv4-pool`        | 10.72.16.0/20
`bmc-ipv4-range-size`  | 5
`bmc-ipv4-range-mask`  | 20

For a node whose rack number is `0` and index in rack is `4`,
its static addresses for node OS are:

* 10.69.0.4
* 10.69.0.68
* 10.69.0.132

and its BMC address is:

* 10.72.17.4

For another node whose rack number is `1` and index in rack is `5`,
its static addresses for node OS are:

* 10.69.0.197
* 10.69.1.5
* 10.69.1.69

and its BMC address is:

* 10.72.17.37

REST API
--------

### Common status code

- JSON parse failure: 400 Bad Request
- `sabakan` internal error: 500 Internal Server Error

### `PUT /api/v1/config/ipam`

Create or update IPAM configurations.  If one or more nodes have been registered in sabakan, IPAM configurations cannot be updated.

**Successful response**

- HTTP status code: 200 OK

**Failure responses**

- One or more nodes are already registered.

  HTTP status code: 500 Internal Server Error

```console
$ curl -XPUT localhost:8888/api/v1/config -d '
{
   "max-nodes-in-rack": 28,
   "node-ipv4-pool": "10.69.0.0/16",
   "node-ipv4-range-size": 6,
   "node-ipv4-range-mask": 26,
   "node-ip-per-node": 3,
   "node-index-offset": 3,
   "bmc-ipv4-pool": "10.72.16.0/20",
   "bmc-ipv4-range-size": 5,
   "bmc-ipv4-range-mask": 20
}'
```

### `GET /api/v1/config/ipam`

Get IPAM configurations.

**Successful response**

- HTTP status code: 200 OK
- HTTP response header: `application/json`
- HTTP response body: JSON

**Failure responses**

- `/<prefix>/config` does not exist in etcd

  HTTP status code: 404 Not Found

```console
$ curl -XGET localhost:8888/api/v1/config
{
   "max-nodes-in-rack": 28,
   "node-ipv4-pool": "10.69.0.0/16",
   "node-ipv4-range-size": 6,
   "node-ipv4-range-mask": 26,
   "node-ip-per-node": 3,
   "node-index-offset": 3,
   "bmc-ipv4-pool": "10.72.16.0/20",
   "bmc-ipv4-range-size": 5,
   "bmc-ipv4-range-mask": 20
}
```

### `POST /api/v1/machines`

Register machines. Sabakan automatically set the `status` to `running,` and `index-in-rack` which is the index number of its machine in the rack and IP addresses. All of the machines in the requested JSON is an atomic operation to register. If Sabakan fails to register at least one machine, it all fails. In other words, the result will be registered all machines or not registered at all. There is no possibility that part of machines will be registered.

In the HTTP request body, specify the following list of the machine information in JSON format.

Field                        | Description
-----                        | -----------
`serial=<serial>`            | The serial number of the machine
`datacenter=<datacenter>`    | The data center name where the machine is in
`rack=<rack>`                | The rack number where the machine is in. If it is omitted, value set to `0`
`role=<role>`                | The role of the machine(`boot` or `worker`)
`product=<product>`          | The product name of the machine(e.g. `R630`)

**Successful response**

- HTTP status code: 201 Created

**Failure responses**

- The same serial number of the machine is already registered.

  HTTP status code: 409 Conflict

- The boot server in the specified `rack` is already registered.

  HTTP status code: 409 Conflict

```console
$ curl -i -X POST \
   -H "Content-Type:application/json" \
   -d \
'[{
  "serial": "1234abcd",
  "product": "R630",
  "datacenter": "ty3",
  "rack": 1,
  "role": "boot"
}]' \
 'http://localhost:8888/api/v1/machines'
```

### `GET /api/v1/machines`

Search registered machines. A user can specify the following URL queries.

Query                      | Description
-----                      | -----------
`serial=<serial>`          | The serial number of the machine
`datacenter=<datacenter>`  | The data center name where the machine is in
`rack=<rack>`              | The rack number where the machine is in
`role=<role>`              | The role of the machine
`index-in-rack=<rack>`     | The unique index number of the machine. It is not relevant with the physical location
`product=<product>`        | The product name of the machine(e.g. `R630`)
`ipv4=<ip address>`        | IPv4 address
`ipv6=<ip address>`        | IPv6 address

**Successful response**

- HTTP status code: 200 OK
- HTTP response header: `application/json`
- HTTP response body: Machines of an array of the JSON

**Failure responses**

- No such machines found.

  HTTP status code: 404 Not Found

```console
$ curl -XGET 'localhost:8888/api/v1/machines?serial=1234abcd'
[{"serial":"1234abcd","product":"R630","datacenter":"us","rack":1,"index-in-rack":1,"role":"boot","network":{"node0":{"ipv4":["10.69.0.197"],"ipv6":null},"node1":{"ipv4":["10.69.1.5"],"ipv6":null},"node2":{"ipv4":["10.69.1.69"],"ipv6":null}},"bmc":{"ipv4":["10.72.17.37"]}}]
$ curl -XGET 'localhost:8888/api/v1/machines?datacenter=ty3&rack=1&product=R630'
[{"serial":"10000000","product":"R630","datacenter":"ty3","rack":1,"index-in-rack":1,"role":"boot","network":{"node0":{"ipv4":["10.69.0.197"],"ipv6":null},"node1":{"ipv4":["10.69.1.5"],"ipv6":null},"node2":{"ipv4":["10.69.1.69"],"ipv6":null}},"bmc":{"ipv4":["10.72.17.37"]}},{"serial":"10000001","product":"R630","datacenter":"ty3","rack":1,"index-in-rack":1,"role":"boot","network":{"node0":{"ipv4":["10.69.0.197"],"ipv6":null},"node1":{"ipv4":["10.69.1.5"],"ipv6":null},"node2":{"ipv4":["10.69.1.69"],"ipv6":null}},"bmc":{"ipv4":["10.72.17.37"]}}]
$ curl -XGET 'localhost:8888/api/v1/machines?ipv4=10.20.30.40'
[{"serial":"20000000","product":"R630","datacenter":"us","rack":1,"index-in-rack":1,"role":"boot","network":{"node0":{"ipv4":["10.69.0.197"],"ipv6":null},"node1":{"ipv4":["10.20.30.40"],"ipv6":null},"node2":{"ipv4":["10.69.1.69"],"ipv6":null}},"bmc":{"ipv4":["10.72.17.37"]}}]
```

### `DELETE /api/v1/machines/<serial>`

Delete registered machine of the `<serial>`.

**Successful response**

- HTTP status code: 200 OK
- HTTP response header: `application/json`
- HTTP response body: JSON

**Failure responses**

- No specified machine found.

  HTTP status code: 404 Not Found

```console
$ curl -i -X DELETE 'localhost:8888/api/v1/machines/1234abcd'
(No output in stdout)
```

### `GET /api/v1/ignitions/<serial>`

Get CoreOS ignition.

```console
$ curl -XGET localhost:8888/api/v1/ignitions/1234abcd
```
!!! Caution
    Not implemented.

### `PUT /api/v1/crypts/<serial>/<path>`

Register disk encryption key. The request body is raw binary format of the key.

**Successful response**

- HTTP status code: 201 Created
- HTTP response header: `application/json`
- HTTP response body: JSON

```json
{"status": 201, "path": <path>}
```

**Failure responses**

- `/<prefix>/crypts/<serial>/<path>` already exists.

  HTTP status code: 409 Conflict

```console
$ echo "binary key data" | curl -i -X PUT -d - \
   'http://localhost:8888/api/v1/crypts/1/aaaaa'
HTTP/1.1 201 Created
Content-Type: application/json
Date: Tue, 10 Apr 2018 09:12:12 GMT
Content-Length: 31

{"status": 201, "path":"aaaaa"}
```

### `GET /api/v1/crypts/<serial>/<path>`

Get an encryption key of the particular disk.

**Successful response**

- HTTP status code: 200 OK
- HTTP response header: `application/octet-stream`
- HTTP response body: A raw key data

**Failure responses**

- No specified `/<prefix>/crypts/<serial>/<path>` found in etcd.

  HTTP status code: 404 Not Found

```console
$ curl -i -X GET \
   'http://localhost:8888/api/v1/crypts/1/aaaaa'
HTTP/1.1 200 OK
Content-Type: application/octet-stream
Date: Tue, 10 Apr 2018 09:15:59 GMT
Content-Length: 64

.....
```

### `DELETE /api/v1/crypts/<serial>`

Delete all disk encryption keys of the specified machine. This request does not delete `/api/v1/machines/<serial>`, User can re-register encryption keys using `<serial>`.

**Successful response**

- HTTP status code: 200 OK
- HTTP response header: `application/json`
- HTTP response body: Array of the `<path>` which are deleted successfully.

```console
$ curl -i -X DELETE \
   'http://localhost:8888/api/v1/crypts/1'
HTTP/1.1 200 OK
Content-Type: application/json
Date: Tue, 10 Apr 2018 09:19:01 GMT
Content-Length: 18

["abdef", "aaaaa"]
```

## `sabactl`

### Usage

```console
$ sabactl [--server http://localhost:8888] <subcommand> <args>...
```

Option     | Default value           | Description
------     | -------------           | -----------
`--server` | `http://localhost:8888` | URL of sabakan

### `sabactl ipam set`

Set/update IPAM configurations.

```console
$ sabactl ipam set -f <sabakan_configurations.json>
```

### `sabactl ipam get`

Get IPAM configurations.

```console
$ sabactl iapm get
```

### `sabactl machines create`

Register new machines.

```console
$ sabactl machines create -f <machine_informations.json>
```

You can register multiple machines by giving a list of machine specs as shown below.
Detailed specification of the input JSON file is same as that of the `POST /api/v1/machines` API.

```json
[
  { "serial": "<serial1>", "datacenter": "<datacenter1>", "rack": <rack1>, "product": "<product1>", "role": "<role1>" },
  { "serial": "<serial2>", "datacenter": "<datacenter2>", "rack": <rack2>, "product": "<product2>", "role": "<role2>" }
]
```

### `sabactl machines get`

Show machines filtered by query parameters.

```console
$ sabactl machines get [--serial <serial>] [--state <state>] [--datacenter <datacenter>] [--rack <rack>] [--product <product>] [--ipv4 <ip address>] [--ipv6 <ip address>]
```

Detailed specification of the query parameters and the output JSON content is same as those of the `GET /api/v1/machines` API.

!!! Note
    `--state <state>` will not be implemented until the policy of machines life-cycle management is fixed.

### `sabactl machines remove`

Unregister a machine specified by a serial number.

```console
$ sabactl machines remove <serial>
```

!!! Note
    This will be refined for machines life-cycle management.
    We should not unregister machines by their serials, but by their statuses.
    We can unregister machines only if their statuses are "to be repaired" or "to be discarded" or anythin like those.
    So the parameters of this command should be `--state <state>`.

## etcd のスキーマ設計

以下のキー/バリューをetcdに作成する。

### `<prefix>/machines/<serial>`


- prefix:   sabakan などの文字列
- serial:   機材のシリアル番号

各機材の情報。データはJSON。

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
    "ipv4": [
      "10.72.17.37"
    ]
  }
```

Key              | Description
---              | -----------
`serial`         | 機材のシリアル番号
`product`        | 機材の製品名(R630等)
`datacenter`     | 機材が置かれているデータセンター
`rack`           | 機材が置かれているラックの論理ラック番号(LRN)
`index-in-rack`  | ラック内の機材を一意に示すインデックス(物理的な場所とは無関係)
`network`        | NIC名がKeyでIPアドレスがvalueの辞書
`bmc`            | BMC(iDRAC)のIPアドレス

### `<prefix>/crypts/<serial>/<path>`

Name   | Description
----   | -----------
prefix | sabakan などの文字列
serial | 機材のシリアル番号
path   | `/dev/disk/by-path` の下のファイル名

機材の各ディスクの暗号鍵。
データは生バイナリの鍵データ。

```console
$ etcdctl get /sabakan/crypts/1234abcd/pci-0000:00:1f.2-ata-3 --print-value-only
(バイナリ鍵)
```

### `<prefix>/ipam`

IPAMの設定。JSON 形式の `sabakan.IPAMConfig`.

### `<prefix>/node-indices/<rack>`

* rack: ラックの番号

ラックごとのノード割り当て情報を登録する。
割り当てたノードインデックスのリストを値とする。

例:
```
$ etcdctl get "/sabakan/node-indices/0"
[3, 4, 5]
```

[BMC]: https://en.wikipedia.org/wiki/Intelligent_Platform_Management_Interface#Baseboard_management_controller
[HTTPBoot]: https://github.com/tianocore/tianocore.github.io/wiki/HTTP-Boot
[Ignition]: https://coreos.com/ignition/docs/latest/
[CoreOS]: https://coreos.com/os/docs/latest/
