sabakan-cryptsetup
==================

`sabakan-cryptsetup` is a utility to help automatic full disk encryption.

It generates disk encryption key and setup encrypted disks by using `cryptsetup`,
a front-end tool of `dm-crypt` kernel module.
The generated encryption key is encrypted with another key and sent to sabakan server.
At the next boot, `sabakan-cryptsetup` will download the encrypted disk encryption key
from sabakan, decrypt it, and setup encrypted disks automatically.

If the server supports [TPM] 2.0, `sabakan-cryptsetup` uses `/dev/tpm0`
to store divided disk encryption key.

If `sabakan-cryptsetup` finds that the disk has been encrypted without TPM information
while the server supports TPM 2.0, it reformats the disk.
When you start using a server, you should enable the TPM of that server at the very beginning.

Usage
-----

```console
$ sabakan-cryptsetup [flags]
```

| Option       | Default value            | Description                      |
| ------------ | ------------------------ | -------------------------------- |
| `--cipher`   | `aes-xts-plain64`        | Cipher specification             |
| `--excludes` | ""                       | Disk name patterns to be ignored |
| `--keysize`  | 512                      | Key size in bits                 |
| `--server`   | `http://localhost:10080` | URL of sabakan                   |
| `--tpmdev`   | `/dev/tpm0`              | TPM character device file        |

| Environment variable | Default value | Description                                  |
| -------------------- | ------------- | -------------------------------------------- |
| `SABAKAN_URL`        | ""            | Default sabakan URL `--server` is not given. |

Target disks
------------

`sabakan-cryptsetup` scans `/sys/block` directory and encrypts all disks excluding:

* Virtual devices (devices not having `/sys/block/*/device`)
* Removable devices (devices whose `/sys/block/*/removable` are not `0`)
* Read-only devices (devices whose `/sys/block/*/ro` are not `0`)

`--excludes` can be used to exclude some disks from automatic encryption.
Following example excludes NVMe snd SCSI disks.

```console
$ sabakan-cryptsetup --excludes 'nvme*' --excludes 'sd*'
```

Crypt device name
-----------------

For each `/sys/block/<NAME>` device, a dm-crypt device is created as `/dev/mapper/crypt-<NAME>`.

[TPM]: https://en.wikipedia.org/wiki/Trusted_Computing
