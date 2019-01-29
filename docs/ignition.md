Ignition Controls
=================

[Ignition][] is a provisioning tool for CoreOS Container Linux.

As a network boot server for CoreOS Container Linux, sabakan provides a template
system for Ignition.  For each machine `role`, administrator can upload Ignition
template to sabakan.

Writing templates
-----------------

An ignition template is defined by a YAML file like this:

```yaml
include: ../base/base.yml
passwd: passwd.yml
files:
  - /etc/hostname
systemd:
  - name: chronyd.service
    enabled: true
networkd:
  - 10-node0.netdev
```

Each field in the template YAML corresponds to an item in [ignition configuration](https://coreos.com/ignition/docs/latest/configuration-v2_3.html):

* `include`: Another ignition template to be included.
* `passwd`: A filename in the same directory of the template.  
    The contents should be a YAML encoded ignition's `passwd` object.
* `files`: List of filenames to be provisioned.  
    The file contents are read from files under `files/` sub directory.
* `systemd`: List of systemd unit files to be provisioned.  
    The unit contents are read from files under `systemd/` sub directory.
* `networkd`: List of networkd unit files to be provisioned.  
    The unit contents are read from files under `networkd/` sub directory.

### Rendering specifications

Files pointed by the template YAML are rendered by [text/template][].

`.` in the template is set to the [`Machine`](machine.md#machine-struct) struct of the target machine.  
For example, `{{ .Spec.Serial }}` will be replaced with the serial number of the target machine.

Following additional template functions are defined and can be used:

* `MyURL`: returns the URL of the sabakan server.
* `Metadata`: takes a key to retrieve metadata value saved along with the template.
* `json`: renders the argument as JSON.

Uploading templates to sabakan
------------------------------

`sabactl ignitions set -f YAML ROLE ID` will upload an ignition template `YAML` for
machines whose role is `ROLE`.  `ID` is a version string such as `1.2.3`.

Sabakan keeps up to 10 old ignition templates for each role.  
When a new ignition template has some defects, administrators can revert it to old one
by deleting the new template.

Templates may have meta data.  To associate meta data, use `--meta` as follows:

```console
$ sabactl ignitions set --meta KEY1=VALUE1,KEY2=VALUE2 -f cs.yaml cs 1.0.0
```

The meta data can be referenced by text/template function `Metadata` as described above.

Example
-------

[`testdata/`](../testdata) directory contains a complete set of ignition templates.

[Ignition]: https://coreos.com/ignition/docs/latest/
[text/template]: https://golang.org/pkg/text/template/
