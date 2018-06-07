sabakan
=======

Usage
-----

```console
$ sabakan -h
Usage of /home/ymmt/go/bin/sabakan:
  -advertise-url string
        public URL of this server
  -config-file string
        path to configuration file
  -dhcp-bind string
        bound ip addresses and port for dhcp server (default "0.0.0.0:10067")
  -etcd-password string
        password for etcd authentication
  -etcd-prefix string
        etcd prefix (default "/sabakan")
  -etcd-servers string
        comma-separated URLs of the backend etcd (default "http://localhost:2379")
  -etcd-timeout string
        dial timeout to etcd (default "2s")
  -etcd-username string
        username for etcd authentication
  -http string
        <Listen IP>:<Port number> (default "0.0.0.0:10080")
  -image-dir string
        directory to store boot images (default "/var/lib/sabakan")
  -ipxe-efi-path string
        path to ipxe.efi (default "/usr/lib/ipxe/ipxe.efi")
  -logfile string
        Log filename
  -logformat string
        Log format [plain,logfmt,json]
  -loglevel string
        Log level [critical,error,warning,info,debug]
```

Option          | Default value            | Description
--------------- | ------------------------ | -----------
`advertise-url` | ""                       | Public URL to access this server.  Required.
`config-file`   | ""                       | If given, configurations are read from the file.
`dhcp-bind`     | `0.0.0.0:10067`          | bound ip addresses and port dhcp server
`etcd-prefix`   | `/sabakan`               | etcd prefix
`etcd-servers`  | `http://localhost:2379`  | comma-separated URLs of the backend etcd
`etcd-timeout`  | `2s`                     | dial timeout to etcd
`etcd-username` | ""                       | username for etcd authentication
`etcd-password` | ""                       | password for etcd authentication
`http`          | `0.0.0.0:10080`          | Listen IP:Port number
`image-dir`     | `/var/lib/sabakan`       | Directory to store boot images.
`ipxe-efi-path` | `/usr/lib/ipxe/ipxe.efi` | path to ipxe.efi

Config file
-----------

Sabakan can read configurations from a YAML file if `-config-file` option is specified.
When `-config-file` is specified, command line options are ignored except for logging
options.

Properties in YAML are the same as the command-line option names without leading slashes.
`etcd-servers` value is a list of URL strings.
