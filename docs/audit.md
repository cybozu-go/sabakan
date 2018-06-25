Audit log
=========

Sabakan records important operations in etcd for audit.
They can be viewed by `sabactl log` sub command.

Log structure
-------------

Each operation log is structured to have following fields:

Field      | Type   | Description
---------- | ------ | -----------
`ts`       | string | The timestamp of the event in [RFC3339][] format.
`rev`      | string | etcd revision of the event.  This is a string-formatted integer.
`user`     | string | UNIX user name who executed `sabactl`.
`ip`       | string | IP address of the host that connected to `sabakan`.
`host`     | string | Hostname where `sabakan` server did the operation.
`category` | string | Operation category such as `machines`, `ipam`, `crypts`, etc.
`instance` | string | ID of the object that was the target of the operation.
`action`   | string | A short verb such as `delete` or `update`.
`detail`   | string | A detailed explanation of the operation.

Compaction
----------

Log entries are kept for 60 days in etcd.  Logs older than 60 days
are automatically removed.  To keep them longer, administrators should
export logs using `sabactl log`.

Note that etcd is not designed to store large objects.  The default
maximum database size is only 2 GiB.

[RFC3339]: https://www.ietf.org/rfc/rfc3339.txt
