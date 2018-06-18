Machine
=======

Machine represents a physical server.

Registration format
-------------------

A machine can be defined by JSON object having these fields:

Field           | Type   | Description
--------------- | ------ | -----------
`serial`        | string | Serial number of the machine.
`product`       | string | Product name of the machine
`datacenter`    | string | Data center name where the machine exists.
`rack`          | int    | Logical rack number (LRN) where the machine exists.
`bmc`           | object | BMC parameters; see below.

Key in `bmc`    | Type   | Description
--------------- | ------ | -----------
`type`          | string | BMC type e.g. "iDRAC-9", "IPMI-2.0".

Response format
---------------

When sabakan registers a machine, it allocates a slot in the rack and adds
it as `index-in-rack` field.  It also allocates IP addresses as configured
by [IPAMConfig](ipam.md#ipamconfig).

Each machine also has `state` as described in [lifecycle management](lifecycle.md).
At registration, the initial state is set to `healthy`.

As a result, a machine have these additional fields in addition to fields
for registration.

Field           | Type     | Description
--------------- | -------- | -----------
`index-in-rack` | int      | Logical position in a rack.
`ipv4`          | []string | IPv4 addresses for OS.
`ipv6`          | []string | IPv6 addresses for OS.
`state`         | string   | The state of the machine.

Key in `bmc`    | Type   | Description
--------------- | ------ | -----------
`ipv4`          | string | BMC's IPv4 address
`ipv6`          | string | BMC's IPv6 address

Example
-------

```json
{
  "serial": "1234abcd",
  "product": "Dell R630",
  "datacenter": "tokyo1",
  "rack": 1,
  "index-in-rack": 1,
  "role": "boot",
  "ipv4": ["10.69.0.69", "10.69.0.133"],
  "ipv6": [],
  "state": "healthy",
  "bmc": {
    "type": "iDRAC-9",
    "ipv4": "10.72.17.37"
  }
}
```
