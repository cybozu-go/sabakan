Machine
=======

```json
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
    "type": "iDRAC-9",
    "ipv4": "10.72.17.37"
  }
```

Key              | Description
---              | -----------
`serial`         | Serial number of the machine
`product`        | Product name of the machine
`datacenter`     | Data center name where the machine is in
`rack`           | Logical rack number (LRN) where the machine is in
`index-in-rack`  | Index number of the machine in a rack; this does not correspond to physical position
`network`        | IP addresses of the machine indexed with NIC names and protocol names (IPv4/IPv6)
`bmc`            | Machine's BMC specs; see below
`state`          | See [Machine Lifecycle Management](lifecycle.md)

Key in `bmc`    | Description
------------    | -----------
`type`          | BMC type e.g. 'iDRAC-9', 'IPMI-2.0'
`ipv4`          | IPv4 address of BMC
`ipv6`          | IPv6 address of BMC
