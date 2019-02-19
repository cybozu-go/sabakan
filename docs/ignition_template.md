Ignition Templates
==================

[Ignition][] is a provisioning tool for CoreOS Container Linux.

As a network boot server for CoreOS Container Linux, sabakan provides a template
system for Ignition.  For each machine `role`, administrator can upload Ignition
template to sabakan.

Writing templates
-----------------

An ignition template is defined by a YAML file like this:

```yaml
version: "2.3"
include: ../base/base.yml
passwd: passwd.yml
files:
  - /etc/hostname
remote_files:
  - name: /tmp/some.img
    url: "{{ MyURL }}/v1/assets/some.img"
systemd:
  - name: chronyd.service
    enabled: true
    mask: false
networkd:
  - 10-node0.netdev
```

Field descriptions:

* `version`: Ignition specification version.  Current supported versions are:
    * "2.2" (default)
    * "2.3"
* `include`: Another ignition template to be included.
* `passwd`: A YAML filename that contains YAML encoded ignition's [`passwd` object](https://coreos.com/ignition/docs/latest/configuration-v2_3.html).
* `files`: List of filenames to be provisioned.  
    The file contents are read from files under `files/` sub directory.
* `remote_files`: List of remote files to be provisioned.
* `systemd`: List of systemd unit files to be provisioned.  
    The unit contents are read from files under `systemd/` sub directory.  
    If `enabled` is true, the unit will be enabled.  
    If `mask` is true, the unit will be masked.
* `networkd`: List of networkd unit files to be provisioned.  
    The unit contents are read from files under `networkd/` sub directory.

### Rendering specifications

Files pointed by the template YAML are rendered by [text/template][].
Strings in `passwd` YAML and `url` for `remote_files` are also rendered as templates.

`.` in the template is set to the [`Machine`](machine.md#machine-struct) struct of the target machine.  
For example, `{{ .Spec.Serial }}` will be replaced with the serial number of the target machine.

Following additional template functions are defined and can be used:

* `MyURL`: returns the URL of the sabakan server.
* `Metadata`: takes a key to retrieve metadata value saved along with the template.
* `json`: renders the argument as JSON.
* `add`, `sub`, `mul`, `div`: do arithmetic on parameters.

For example, the following template may be replaced with 6 when `Machine.Spec.Rack` is 3.

```
{{ add .Spec.Rack 3 }}
```

Uploading templates to sabakan
------------------------------

`sabactl ignitions set -f YAML ROLE ID` will upload an ignition template `YAML` for
machines whose role is `ROLE`.  `ID` is a version string such as `1.2.3`.

Sabakan keeps up to 10 old ignition templates for each role.  
When a new ignition template has some defects, administrators can revert it to old one
by deleting the new template.

Templates may have meta data.  To associate meta data, create a JSON file
containing a JSON object, and specify it with `--meta FILENAME` as follows:

```console
$ cat >meta.json <<'EOF'
{
  "key1": "value1",
  "key2": [1, 2, 3]
}
EOF
$ sabactl ignitions set --meta meta.json -f cs.yaml cs 1.0.0
```

The meta data can be referenced by [text/template][] function `Metadata` as described above.

Example
-------

[`testdata/`](../testdata) directory contains a complete set of ignition templates.

[Ignition]: https://coreos.com/ignition/docs/latest/
[text/template]: https://golang.org/pkg/text/template/
