sabakan-cryptsetup
==================

Usage
-----

```console
$ sabakan-cryptsetup [--server http://localhost:10080] DISK_PATH_PATTERN...
```

Option     | Default value            | Description
------     | -------------            | -----------
`--server` | `http://localhost:10080` | URL of sabakan


Environment variable | Default value | Description
-------------------- | ------------- | -----------
`SABAKAN_URL`        | ""            | Default sabakan URL when commands don't specify `--server`

`sabakan-cryptsetup DISK_PATH_PATTERN`
--------------------------------------

Prepare encryption disks which are detected on the system.
`DISK_PATH_PATTERN` finds device file path under `/dev/disk/by-path`.
Detected disks will be encrypted and/or decrypted using keys where store on `sabakan`. See [details](disk_encryption.md).

```console
$ sabakan-cryptsetup virtio-pci-*
```
