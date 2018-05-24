sabakan
=======

Usage
-----

```console
$ sabakan [-dhcp-bind 0.0.0.0:10067] [-etcd-servers http://foo.bar:2379,http://zot.bar:2379] ...
```

Option           | Default value            | Description
------           | -------------            | -----------
`-dhcp-bind`     | `0.0.0.0:10067`          | bound ip addresses and port dhcp server
`-etcd-prefix`   | `/sabakan`               | etcd prefix
`-etcd-servers`  | `http://localhost:2379`  | comma-separated URLs of the backend etcd
`-etcd-timeout`  | `2s`                     | dial timeout to etcd
`-http`          | `0.0.0.0:10080`          | Listen IP:Port number
`-ipxe-efi-path` | `/usr/lib/ipxe/ipxe.efi` | path to ipxe.efi
`-url-port`      | `10080`                  | port number used to construct boot API URL
