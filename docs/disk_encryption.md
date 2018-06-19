Disk encryption
===============

In order to mount encrypted disk, Sabakan provides `sabakan-cryptsetup` as follows.

* Sabakan keeps encryption keys of all disks which are specified by `sabakan-cryptsetup` argument.

    If at least one key is not registered on Sabakan, all target disks will be re-initialized by new keys.
    `sabakan-cryptsetup` access `/api/v1/crypts` to manage keys. See details [API](api.md).

* Operators need to setup RAID, filesystem format, and disk mount after decrypt disks.

    Operators could prepare other scripts or systemd units to mount disks.

### Procedure of encryption disk mount

1. `sabakan-cryptsetup` makes encrypted device mappers which are specified by its option.
2. Operator setup RAID, and/or format filesystem.
3. Operator mount filesystem.
