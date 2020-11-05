Concepts
========

Node / Machine
--------------

The primary purpose of sabakan is to manage tons of physical servers in a data center.
A physical server is called a *node* or a *machine* in sabakan.

BMC
---

Sabakan assumes each physical server equips a [baseboard management controller][BMC].
BMC is a dedicated hardware to manage the server remotely via network.

Rack
----

Sabakan assumes that nodes are located in *racks*.  A rack can have limited number
of servers.  Sabakan divides a range of network address space into small ranges.
Each divided range is for a rack and should have enough capacity to allocate IP
address to all servers in the rack.

Role
----

Each node has *role* attribute.  Roles can be defined arbitrary as string values,
except for **"boot"**.  Role "boot" is special; only one **boot** node can exist
in a rack.

Node index in rack
------------------

Every node has an index that indicates a logical position in a rack.
No two nodes share the same index if they are in the same rack.

The index is used to calculate IP addresses to be assigned to the node.

The index cannot be specified externally; sabakan sets the index of a node
when the node is registered.

UEFI HTTP Boot
--------------

[UEFI HTTP Boot][HTTPBoot] is a modern network boot technology that replaces
legacy [PXE boot](https://en.wikipedia.org/wiki/Preboot_Execution_Environment).

UEFI HTTP Boot requires DHCP to configure the initial networking.  However,
unlike PXE, HTTP boot does not use TFTP to load boot loaders.  Instead of TFTP,
HTTP is used.

Sabakan is optimized for UEFI HTTP Boot.  It speaks DHCP and HTTP, but does
not speak TFTP.

Ignition
--------

[Ignition][] is a provisioning utility for [Flatcar Container Linux][Flatcar].
Its configuration is just a plain JSON.

Sabakan provides a versatile template system to generate Ignition configuration
for each machine.

[BMC]: https://en.wikipedia.org/wiki/Intelligent_Platform_Management_Interface#Baseboard_management_controller
[HTTPBoot]: https://github.com/tianocore/tianocore.github.io/wiki/HTTP-Boot
[Ignition]: https://coreos.com/ignition/docs/latest/
[Flatcar]: https://docs.flatcar-linux.org/
