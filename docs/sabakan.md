sabakan
=======

Usage
-----

```console
$ sabakan -h
Usage of sabakan:
  -advertise-url string
        public URL of this server
  -allow-ips string
        comma-separated IPs allowed to change resources (default "127.0.0.1,::1")
  -config-file string
        path to configuration file
  -data-dir string
        directory to store files (default "/var/lib/sabakan")
  -dhcp-bind string
        bound ip addresses and port for dhcp server (default "0.0.0.0:10067")
  -etcd-endpoints string
        comma-separated URLs of the backend etcd endpoints (default "http://localhost:2379")
  -etcd-password string
        password for etcd authentication
  -etcd-prefix string
        etcd prefix (default "/sabakan/")
  -etcd-timeout string
        dial timeout to etcd (default "2s")
  -etcd-tls-ca string
        path to CA bundle used to verify certificates of etcd servers
  -etcd-tls-cert string
        path to my certificate used to identify myself to etcd servers
  -etcd-tls-key string
        path to my key used to identify myself to etcd servers
  -etcd-username string
        username for etcd authentication
  -http string
        <Listen IP>:<Port number> (default "0.0.0.0:10080")
  -ipxe-efi-path string
        path to ipxe.efi (default "/usr/lib/ipxe/ipxe.efi")
  -logfile string
        Log filename
  -logformat string
        Log format [plain,logfmt,json]
  -loglevel string
        Log level [critical,error,warning,info,debug]
```

Option           | Default value            | Description
---------------  | ------------------------ | -----------
`advertise-url`  | ""                       | Public URL to access this server.  Required.
`allow-ips`      | `127.0.0.1,::1`          | comma-separated IPs allowed to change resources
`config-file`    | ""                       | If given, configurations are read from the file.
`data-dir`       | `/var/lib/sabakan`       | Directory to store files.
`dhcp-bind`      | `0.0.0.0:10067`          | bound ip addresses and port dhcp server
`etcd-endpoints` | `http://localhost:2379`  | comma-separated URLs of the backend etcd
`etcd-password`  | ""                       | password for etcd authentication
`etcd-prefix`    | `/sabakan`               | etcd prefix
`etcd-timeout`   | `2s`                     | dial timeout to etcd
`etcd-tls-ca`    | ""                       | Path to CA bundle used to verify certificates of etcd endpoints.
`etcd-tls-cert`  | ""                       | Path to my certificate used to identify myself to etcd servers.
`etcd-tls-key`   | ""                       | Path to my key used to identify myself to etcd servers.
`etcd-username`  | ""                       | username for etcd authentication
`http`           | `0.0.0.0:10080`          | Listen IP:Port number
`ipxe-efi-path`  | `/usr/lib/ipxe/ipxe.efi` | path to ipxe.efi

Config file
-----------

Sabakan can read configurations from a YAML file if `-config-file` option is specified.
When `-config-file` is specified, command line options are ignored except for logging
options.

Properties in YAML are the same as the command-line option names without leading slashes.
etcd config is defined by [cybozu-go/etcdutil](https://github.com/cybozu-go/etcdutil)
