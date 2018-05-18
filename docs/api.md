REST API
========

* [PUT /api/v1/config/ipam](#putipam)
* [GET /api/v1/config/ipam](#getipam)
* [POST /api/v1/machines](#postmachines)
* [GET /api/v1/machines](#getmachines)
* [DELETE /api/v1/machines](#deletemachines)
* [GET /api/v1/ignitions](#getignitions)
* [PUT /api/v1/crypts](#putcrypts)
* [GET /api/v1/crypts](#getcrypts)
* [DELETE /api/v1/crypts](#deletecrypts)

## <a name="putipam" />`PUT /api/v1/config/ipam`

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

## <a name="getipam" />`GET /api/v1/config/ipam`

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

## <a name="postmachines" />`POST /api/v1/machines`

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

## <a name="getmachines" />`GET /api/v1/machines`

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

## <a name="deletemachines" />`DELETE /api/v1/machines/<serial>`

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

## <a name="getignitions" />`GET /api/v1/ignitions/<serial>`

Get CoreOS ignition.

```console
$ curl -XGET localhost:8888/api/v1/ignitions/1234abcd
```
!!! Caution
    Not implemented.

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
$ curl -i -X GET \
   'http://localhost:8888/api/v1/crypts/1/aaaaa'
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
$ curl -i -X DELETE \
   'http://localhost:8888/api/v1/crypts/1'
HTTP/1.1 200 OK
Content-Type: application/json
Date: Tue, 10 Apr 2018 09:19:01 GMT
Content-Length: 18

["abdef", "aaaaa"]
```