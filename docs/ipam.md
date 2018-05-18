IP address management (IPAM)
============================

Overview
--------

Sabakan assigns IP addresses to nodes for two purposes.  One is for DHCP, and another is for static assignment.  DHCP is used for the initial network boot.

Each machine may have one or more IP addresses for OS, and one for its [baseboard management controller][BMC].

IPAMConfig
----------

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

Setting the index of a node
---------------------------

Upon node registration, sabakan sets the index in rack of the node.

Nodes whose role is "boot" will have `node-index-offset` as its index in rack.
Other nodes will have an index ranging from `node-index-offset + 1` to `node-index-offset + max-nodes-in-rack`.

Assigning static IPv4 addresses to Node OS
------------------------------------------

Sabakan computes static IPv4 addresses for a node OS as follows (pseudo code):

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

Assigning static IPv4 address to BMC
------------------------------------

Sabakan computes static IPv4 addresses for BMC as follows (pseudo code):

```go
rack := node.RackNumber
idx  := node.IndexInRack

range_size := 1 << bmc-ipv4-range-size
base := INET_ATON(bmc-ipv4-pool)

bmc_addr := INET_NTOA(base + range_size * rack + idx)
```

DHCP lease range
----------------

In short, sabakan leases IP addresses for DHCP that are never assigned
statically to nodes.

Specifically, the range of IP addresses for lease begins from the next
address of the node whose index in rack is `node-index-offset + max-nodes-in-rack`,
and ends at the second last address of the divided range of the IP address pool.

Note that the divided range to be used can be determined by the interface
address that accepts the DHCP request because the interface address must
belong to the range.  If the request was relayed by another DHCP server,
the interface address of the relaying server should be used instead.

Examples
--------

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
