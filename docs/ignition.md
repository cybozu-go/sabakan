Ignition Controls
=================

In order to distribute CoreOS ignitions, sabakan provides an ignition management system as follows.

* Operators need to upload template ignitions to one sabakan server.

    Rest of sabakan servers will automatically save by etcd cluster.

* Sabakan saves ignition templates for each `<role>`.

    `<role>` is referred from a parameter `<role>` in a machine.

* Sabakan keeps versions of ignitions by `<role>`.

    In case of a new ignition fatal detects, the change can be rolled back by DELETE API. 
    When operators put a new ignition template to the sabakan, `<id>` is automatically generated. `<id>` format is timestamp such as `1527731687`.
    Running CoreOS can refer a kernel parameter `coreos.config.url` to know which `<id>` of the ignition template applied.

* Sabakan saves configured number of ignitions.

    Sabakan deletes oldeh ignitions when operators upload a new ignition template.

How it works
------------

### Serving ignitions from CoreOS

1. iPXE received DHCP IP address with iPXE script which kernel parameter includes the newest ignition URL.
2. HTTP BOOT retrieves kernel and initrd. 
3. The initrd download ignition file from sabakan.
4. CoreOS apply downloaded ignition to its system.
