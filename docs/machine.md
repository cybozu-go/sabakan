Machine
=======

Machine represents a physical server.

MachineSpec struct
------------------

`MachineSpec` is a struct to define a Machine.

`MachineSpec` can be represented as a JSON object having these fields:

Field           | Type     | Description
--------------- | -------- | -----------
`serial`        | string   | SMBIOS serial number of the machine.
`product`       | string   | Product name of the machine
`datacenter`    | string   | Data center name where the machine exists.
`rack`          | int      | Logical rack number (LRN) where the machine exists.
`index-in-rack` | int      | Logical position in a rack.
`ipv4`          | []string | IPv4 addresses for OS.
`ipv6`          | []string | IPv6 addresses for OS.
`bmc`           | object   | BMC parameters; See below.

Key in `bmc`    | Type   | Description
--------------- | ------ | -----------
`type`          | string | BMC type e.g. "iDRAC-9", "IPMI-2.0".
`ipv4`          | string | BMC's IPv4 address
`ipv6`          | string | BMC's IPv6 address

Specifically, `index-in-rack`, `ipv4`, `ipv6` `bmc.ipv4`, and `bmc.ipv4` fields
are automatically filled by sabakan at registration.

Machine struct
--------------

`Machine` has `MachineSpec` and its status as described in [lifecycle management](lifecycle.md).

A JSON representation of `Machine` looks like:

```json
{
  "spec": {
    "serial": "1234abcd",
    "product": "Dell R630",
    "datacenter": "tokyo1",
    "rack": 1,
    "index-in-rack": 1,
    "role": "boot",
    "ipv4": ["10.69.0.69", "10.69.0.133"],
    "ipv6": [],
    "bmc": {
      "type": "iDRAC-9",
      "ipv4": "10.72.17.37"
    }
  },

  "status": {
    "state": "healthy",
    "timestamp": "2018-06-19T03:43:25.46669721Z"
  }
}
```
