DHCP
====

Sabakan works as a DHCP server.

There are a few configuration options for the DHCP service.

DHCPConfig
----------

`DHCPConfig` is a set of configurations for DHCP options.
It is given as a JSON object with the following fields:

Field            | Required | Type            | Description
---------------- | -------- | --------------- | -----------
`lease-minutes`  | No       | int             | Lease period in minutes.  Default is 60.
`dns-servers`    | No       | array of string | The IP addresses of DNS servers.
