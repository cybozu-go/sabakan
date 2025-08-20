sabakan
=======

Usage
-----

See [specification of etcdutil](https://github.com/cybozu-go/etcdutil/blob/main/README.md#specifications) for etcd connection flags and parameters.

```console
$ sabakan -h
Usage of sabakan:
  -advertise-url string
        public URL of this server
  -advertise-url-https string
        public URL of this server(https)
  -allow-ips string
        comma-separated IPs allowed to change resources (default "127.0.0.1,::1")
  -config-file string
        path to configuration file
  -data-dir string
        directory to store files (default "/var/lib/sabakan")
  -dhcp-bind string
        bound ip addresses and port for dhcp server (default "0.0.0.0:10067")
  -enable-playground
        enable GraphQL playground
  -etcd-endpoints string
        comma-separated URLs of the backend etcd endpoints (default "http://127.0.0.1:2379")
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
  -https string
        <Listen IP>:<Port number> (default "0.0.0.0:10443")
  -ipxe-efi-path string
        path to ipxe.efi (default "/usr/lib/ipxe/ipxe.efi")
  -logfile string
        Log filename
  -logformat string
        Log format [plain,logfmt,json]
  -loglevel string
        Log level [critical,error,warning,info,debug]
  -metrics string
        <Listen IP>:<Port number> (default "0.0.0.0:10081")
  -metrics-leader-election-ttl string
        TTL for the etcd session used by metrics leader election. Accepts Go duration (e.g., 30s, 60s) (default "60s")
  -server-cert string
        path to server TLS certificate of sabakan (default "/etc/sabakan/server.crt")
  -server-key string
        path to server TLS key of sabakan (default "/etc/sabakan/server.key")
```

| Option                       | Default value                      | Description                                                     |
| ---------------------------- | ---------------------------------- | --------------------------------------------------------------- |
| `advertise-url`              | ""                                 | Public URL to access HTTP server.  Required.                    |
| `advertise-url-https`        | ""                                 | Public URL to access HTTPS server.  Required.                   |
| `allow-ips`                  | `127.0.0.1,::1`                    | Comma-separated IPs allowed to change resources.                |
| `config-file`                | ""                                 | If given, configurations are read from the file.                |
| `data-dir`                   | `/var/lib/sabakan`                 | Directory to store files.                                       |
| `dhcp-bind`                  | `0.0.0.0:10067`                    | IP address and port number of DHCP server.                      |
| `enable-playground`          | false                              | Enable GraphQL playground service.                              |
| `etcd-endpoints`             | `http://127.0.0.1:2379`            | Comma-separated URLs of the backend etcd endpoints.             |
| `etcd-password`              | ""                                 | Password for etcd authentication.                               |
| `etcd-prefix`                | `/sabakan/`                        | etcd prefix.                                                    |
| `etcd-timeout`               | `2s`                               | Dial timeout to etcd.                                           |
| `etcd-tls-ca`                | ""                                 | Path to CA bundle used to verify certificates of etcd servers.  |
| `etcd-tls-cert`              | ""                                 | Path to my certificate used to identify myself to etcd servers. |
| `etcd-tls-key`               | ""                                 | Path to my key used to identify myself to etcd servers.         |
| `etcd-username`              | ""                                 | Username for etcd authentication.                               |
| `http`                       | `0.0.0.0:10080`                    | IP address and port number of HTTP server.                      |
| `https`                      | `0.0.0.0:10443`                    | IP address and port number of HTTPS server.                     |
| `ipxe-efi-path`              | `/usr/lib/ipxe/ipxe.efi`           | Path to ipxe.efi .                                              |
| `metrics`                    | `0.0.0.0:10081`                    | IP address and port number of metrics HTTP server.              |
| `metrics-leader-election-ttl`| `60s`                              | TTL for the etcd session used by metrics leader election.       |
| `server-cert`                | `/etc/sabakan/server.crt`          | Path to server  certificate of sabakan.                         |
| `server-key`                 | `/etc/sabakan/server.key`          | Path to server TLS key of sabakan.                              |

Config file
-----------

Sabakan can read configurations from a YAML file if `-config-file` option is specified.
When `-config-file` is specified, command line options are ignored except for logging
options.

Properties in YAML are the same as the command-line option names without leading slashes.
etcd config can be defined `etcd:`. The etcd parameters are defined by [cybozu-go/etcdutil](https://github.com/cybozu-go/etcdutil), and not shown below will use default values of the etcdutil.

| Name     | Type   | Required | Description                                          |
| -------- | ------ | -------- | ---------------------------------------------------- |
| `prefix` | string | No       | Key prefix of etcd objects.  Default is `/sabakan/`. |

Environment variable
--------------------

| Name                 | Description                              |
| -------------------- | ---------------------------------------- |
| `SABAKAN_CRYPTSETUP` | Path to `sabakan-cryptsetup` executable. |

* If `SABAKAN_CRYPTSETUP` is not specified, `sabakan-cryptsetup` will be looked up
    in the same directory of `sabakan` executable file.
