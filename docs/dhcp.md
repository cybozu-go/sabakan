DHCP
====

Overview
--------

Sabakan speaks DHCP mainly for UEFI HTTP boot.
As DHCP can offer information such as addresses of the default gateway,
NTP server, or DNS server, sabakan has configuration options for them.

DHCPConfig
----------

`DHCPConfig` is a set of configurations for DHCP options.
It is given as JSON object with the following fields:

Field            | Required | Type   | Description
---------------- | -------- | ------ | -----------
`gateway-offset` | Yes      | int    | The default gateway address offset.  See below.

### `gateway-offset`

A default gateway need to be given to clients to communicate with
servers outside of the L2 network.  The default gateway address is
henceforth an address in the L2 network.

To calculate it automatically, sabakan uses `gateway-offset` as follows.
Suppose the client is receiving "10.2.3.4/24",

1. Mask the address with the netmask to obtain "10.2.3.0"
2. Add `gateway-offset` value; if it is 1, then the result becomes "10.2.3.1".
3. "10.2.3.1" is the default gateway address.
