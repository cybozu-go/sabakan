sabakan-cryptsetup
==================

`sabakan-cryptsetup` is a utility to help automatic full disk encryption.

It generates disk encryption key and setup encrypted disks by using `cryptsetup`,
a front-end tool of `dm-crypt` kernel module.
The generated encryption key is encrypted with another key and sent to sabakan server.
At the next boot, `sabakan-cryptsetup` will download the encrypted disk encryption key
from sabakan, decrypt it, and setup encrypted disks automatically.

If `sabakan-cryptsetup` server supports [TPM] 2.0, `sabakan-cryptsetup` uses `/dev/tpm0`
as a key for generating disk encryption key.

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

`sabakan-cryptsetup` scans `/sys/block` directory and encrypt all disks excluding:

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

Disk layout
-----------

Disks encrypted with `sabakan-cryptsetup` have 2 MiB of meta data at the beginning.
The meta data itself is not encrypted.  The format of meta data is as follows:

| Offset | Length (bytes) | Value                     |
| ------ | -------------: | ------------------------- |
| 0x0000 |             20 | "\x80sabakan-cryptsetup2" |
| 0x0014 |              1 | Key size (bytes)          |
| 0x0015 |              1 | Length of cipher name     |
| 0x0016 |            106 | cipher name               |
| 0x0080 |             16 | Random ID                 |
| 0x0090 |           vary | Key encryption key        |

* The maximum length of cipher name is 106.
* Unused areas are filled with `0x88`.
* The size of key encryption key (KEK) is the same as the key size at 0x0014.

[TPM]: https://en.wikipedia.org/wiki/Trusted_Computing
