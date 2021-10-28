Machine
=======

Machine represents a physical server.

MachineSpec struct
------------------

`MachineSpec` is a struct to define a Machine.

It is represented as a JSON object having these fields:

Field           | Type       | Auto | Description
--------------- | ---------- | ---- | -----------
`serial`        | `string`   | no   | SMBIOS serial number of the machine.
`labels`        | `object`   | no   | `map[string]string` for arbitrary labels.
`rack`          | `int`      | no   | Logical rack number (LRN) where the machine exists.
`index-in-rack` | `int`      | yes  | Logical position in a rack.
`role`          | `string`   | no   | Role of the machine, e.g. `boot`.
`ipv4`          | `[]string` | yes  | IPv4 addresses for OS.
`ipv6`          | `[]string` | yes  | IPv6 addresses for OS.
`register-date` | `string`   | yes  | RFC3339-format date when the machine is registered.
`retire-date`   | `string`   | no   | RFC3339-format date when the machine will be retired.
`bmc`           | `object`   |      | BMC parameters; See below.

Key in `bmc`    | Type     | Auto | Description
--------------- | -------- | ---- | -----------
`type`          | `string` | no   | BMC type e.g. "iDRAC-9", "IPMI-2.0".
`ipv4`          | `string` | yes  | BMC's IPv4 address
`ipv6`          | `string` | yes  | BMC's IPv6 address

Values for auto fields are filled by sabakan at registration.
These auto fields are not accepted in [`sabactl machines create`](docs/sabactl.md#sabactl-machines-create--f-file) because they are overwritten by sabakan.
A partially restricted format of this structure is used for the input values of [`sabactl machines create`](docs/sabactl.md#sabactl-machines-create--f-file).

Machine struct
--------------

`Machine` has `MachineSpec` and its `status` as described in [lifecycle management](lifecycle.md)
as well as additional `info`.

A JSON representation of `Machine` looks like:

```json
{
  "spec": {
    "serial": "1234abcd",
    "labels": {
      "product": "R630",
      "datacenter": "tokyo1"
    },
    "rack": 1,
    "index-in-rack": 1,
    "role": "boot",
    "ipv4": ["10.69.0.69", "10.69.0.133"],
    "ipv6": [],
    "register-date": "2018-11-21T01:02:03.456789Z",
    "retire-date": "2023-11-21T01:02:03.456789Z",
    "bmc": {
      "type": "iDRAC-9",
      "ipv4": "10.72.17.37"
    }
  },

  "status": {
    "state": "healthy",
    "timestamp": "2018-06-19T03:43:25.46669721Z",
    "duration": 362.1
  },

  "info": {
    "network": {
      "ipv4": [
        {
          "address": "10.69.0.69",
          "netmask": "255.255.255.192",
          "maskbits": 26,
          "gateway": "10.69.0.65"
        },
        {
          "address": "10.69.0.133",
          "netmask": "255.255.255.192",
          "maskbits": 26,
          "gateway": "10.69.0.129"
        }
      ]
    },
    "bmc": {
      "ipv4": {
        "address": "10.72.17.37",
        "netmask": "255.255.255.0",
        "maskbits": 24,
        "gateway": "10.71.17.1"
      }
    }
  }
}
```

`status.duration` is the duration between the current time and `status.timestamp`.
The unit of `status.duration` is seconds.

`info.network` contains server NIC configurations.
`info.bmc` contains BMC NIC configuration.
