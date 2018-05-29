sabakan
=======

Usage
-----

```console
$ sabakan -h
Usage of /home/ymmt/go/bin/sabakan:
  -config-file string
        path to configuration file
  -dhcp-bind string
        bound ip addresses and port for dhcp server (default "0.0.0.0:10067")
  -etcd-prefix string
        etcd prefix (default "/sabakan")
  -etcd-servers string
        comma-separated URLs of the backend etcd (default "http://localhost:2379")
  -etcd-timeout string
        dial timeout to etcd (default "2s")
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
  -url-port string
        port number used to construct boot API URL (default "10080")
```

Option            | Default value            | Description
------            | -------------            | -----------
`-config-file` | ""                       | If given, configurations are read from the file.
`-dhcp-bind`    | `0.0.0.0:10067`          | bound ip addresses and port dhcp server
`-etcd-prefix`   | `/sabakan`               | etcd prefix
`-etcd-servers`  | `http://localhost:2379`  | comma-separated URLs of the backend etcd
`-etcd-timeout`  | `2s`                     | dial timeout to etcd
`-http`          | `0.0.0.0:10080`          | Listen IP:Port number
`-ipxe-efi-path` | `/usr/lib/ipxe/ipxe.efi` | path to ipxe.efi
`-url-port`      | `10080`                  | port number used to construct boot API URL

Config file
-----------

Sabakan can read configurations from a YAML file if `-config-file` option is specified.

Logging options can be specified only by command-line options.

Properties in YAML are the same as the command-line option names without leading slashes.
`etcd-servers` value is a list of URL strings.
