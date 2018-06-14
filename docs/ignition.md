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

### Ignition configuration

Sabakan generates ignitions from ignition-like YAML format as a source.
The YAML contains text template to apply machine's parameters and the value of
`storage.files.contents.source` is in plain text but not URL encoded:

```yaml
ignition:
  version: "2.2.0"
storage:
  files:
  - filesystem: root
    path: "/etc/hostname"
    mode: 420
    contents:
      source: "{{ .Serial }}"`,
```

The variables `sabakan` used in YAML are defined in [Machine](https://github.com/cybozu-go/sabakan/blob/d1a01d79307d3b3e188ff7a909204d71b5c2b9bb/machines.go#L12-L22) struct.
The context of the template is the instance of the struct.
For example, the YAML can refers serial number of the machine by `.Serial`.
And the `MyURL` function can be used in YAML. The function will return the url of this sabakan server itself.

User can set YAML files to sabakan via REST API `POST /api/v1/ignitions/ROLE` for each roles.
The iPXE booted machine loads rendered ignitions in JSON via REST API `GET /api/v1/boot/coreos/ipxe/SERIAL`.
Sabakan applies the template parameters to YAML and convert it to JSON format on REST API `GET /api/v1/boot/coreos/ipxe/SERIAL`.

### Setting ignitions by sabactl

`sabactl` supports user-friendly yaml format to generate and register ignitions easily.
It is generated from YAML and source files as the following directory layout:

```text
|- site.yml
|- common.yml
|- files
|   `- etc
|       `- hostname
|- networkd
|   `- 10-eno1.network
`- passwd.yml
```

`site.yml` is an entry point file.  It contains the files to include into the ignition.

```yaml
# site.yml
include: base.yml
passwd: passwd.yml
files:
  - /etc/hostname
systemd:
  - enabled: false
    source: amyapp.service
networkd:
  - 10-eno1.netdev
```

This entry point file contains the following fields:

- `include`: Relative path to extend YAML file.
- `passwd`: Relative path to file which describes user and group settings.
- `files`: String list to deploy static files to Container Linux.
The file content of the item is extract to `storage.files.contents.source` in the ignition.
The path must be *absolute path* and user need to put the source file in the `files` directory.
- `systemd`: Systemd services configuration.  The items are extracted to `systemd` field in the ignition.
    - `enabled`: Enable the service if `true`, otherwise the service will be disabled.
    - `source`: Source file of the systemd.  User need to put the source file in `systemd` directory.
      The basename of the file correspond to the service name.
- `networkd`: Systemd-networkd configuration files.  User need to put the source file in `networkd` directory.

```conf
# files/etc/hostname
{{ .Serial }}
```

```ini
# systemd/myapp.service
[Unit]
Description=Some simple daemon

[Service]
ExecStart=/usr/share/oem/bin/myapp.service --server={{ MyURL }}

[Install]
WantedBy=multi-user.target
```

```ini
# networkd/10-eno1.network
[Match]
Name=eno1

[Network]
Address=10.1.10.10/24
Gateway=10.1.10.1
```

The file described in `passwd` is described in as `passwd` field in the
[Ignition v2.1.0](https://coreos.com/ignition/docs/latest/configuration-v2_1.html).

```yaml
# passwd.yml
users:
  - name: core
    passwordHash: "$6$43y3tkl..."
    sshAuthorizedKeys:
      - key1
groups:
  - name: mgmt
    gid: 10000
```

From the above files, user can register an ignition by `sabactl` command:

```console
# sabactl ignitions set -f <FILE> <ROLE>
$ sabactl ignitions set -f worker.yml worker
```

