Getting started
===============

This document quickly guides you to configure sabakan and netboot
your servers with CoreOS Container Linux.

* [Setup sabakan](#setup)
  * [Prepare etcd](#etcd)
  * [Prepare data directory](#datadir)
  * [Prepare sabakan.yml](#configure)
  * [Run sabakan](#run)
* [Netboot](#netboot)
  * [Configure IPAM](#ipam)
  * [Configure DHCP](#dhcp)
  * [Upload CoreOS Container Linux](#upload)
  * [Register machines](#register)
  * [Register kernel parameters](#kernelparams)
* [What's next](#whatsnext)

## <a name="setup" />Setup sabakan

### <a name="etcd" />Prepare etcd

Sabakan requires [etcd][].  Install and run it at `localhsot`.

You may use docker to run etcd as follows:
```console
$ docker pull quay.io/cybozu/etcd:3.3
$ docker run -d --rm --name etcd --network=host --uts=host quay.io/cybozu/etcd:3.3
```

### <a name="datadir" />Prepare data directory

```console
$ sudo mkdir -p /var/lib/sabakan
```

### <a name="configure" />Prepare sabakan.yml

Save the following contents as `/usr/local/etc/sabakan.yml`:

```yaml
etcd:
  endpoints:
    - http://localhost:2379
dhcp-bind: 0.0.0.0:67
```

For other options, read [sabakan.md](sabakan.md).

### <a name="run" />Run sabakan

Compile and run sabakan as follows:

```console
$ GOPATH=$HOME/go
$ mkdir -p $GOPATH/src
$ export GOPATH
$ go get -u github.com/cybozu-go/sabakan/...
$ sudo $GOPATH/bin/sabakan -config-file /usr/local/etc/sabakan.yml
```

A sample systemd service file is available at
[cmd/sabakan/sabakan.service](../cmd/sabakan/sabakan.service).

Alternatively, you may use docker to run sabakan:
* Repository: [quay.io/cybozu/sabakan](https://quay.io/cybozu/sabakan)
* Usage: https://github.com/cybozu/neco-containers/blob/master/sabakan/README.md

## <a name="netboot" />Netboot

### <a name="ipam" />Configure IPAM

Prepare `ipam.json` as follows:
```json
{
   "max-nodes-in-rack": 28,
   "node-ipv4-pool": "10.69.0.0/20",
   "node-ipv4-range-size": 6,
   "node-ipv4-range-mask": 26,
   "node-index-offset": 3,
   "node-ip-per-node": 3,
   "bmc-ipv4-pool": "10.72.16.0/20",
   "bmc-ipv4-offset": "0.0.1.0",
   "bmc-ipv4-range-size": 5,
   "bmc-ipv4-range-mask": 20
}
```

Then put the JSON to sabakan:
```console
$ sabactl ipam set -f ipam.json
```

Read [ipam.md](ipam.md) for details.

### <a name="dhcp" />Configure DHCP

Prepare `dhcp.json` as follows:
```json
{
   "gateway-offset": 1,
}
```

Then put the JSON to sabakan:
```console
$ sabactl dhcp set -f dhcp.json
```

Read [dhcp.md](dhcp.md) for details.

### <a name="upload" />Upload CoreOS Container Linux

Download CoreOS PXE boot images:
```console
$ curl -o kernel -Lf http://stable.release.core-os.net/amd64-usr/current/coreos_production_pxe.vmlinuz
$ curl -o initrd.gz -Lf http://stable.release.core-os.net/amd64-usr/current/coreos_production_pxe_image.cpio.gz
```

Upload them to sabakan as follows:
```console
$ sabactl images upload ID kernel initrd.gz
```

### <a name="machines" />Register machines

Prepare `machines.json` as follows:
```json
[
  {
    "serial": "1234abcd",
    "labels": {
      "product": "Dell R640",
      "datacenter": "tokyo1"
    },
    "rack": 0,
    "role": "boot",
    "bmc": {
      "type": "IPMI-2.0",
      "ipv4": "10.72.17.37"
    }
  },
  {
    // another machine
  }
]
```

Then put the JSON to sabakan:
```console
$ sabactl machines create -f machines.json
```

Each object in the array is [`MachineSpec`](machine.md#machinespec-struct).
Sabakan identifies physical servers by `serial`.

Once machines are properly registered with sabakan, they can netboot
CoreOS Container Linux using [UEFI HTTP Boot][HTTPBoot].

CoreOS can be initialized at first boot by [ignition][].
Sabakan can generate ignition configuration from templates.
Read [ignition.md](ignition.md) for details.

### <a name="kernelparams" />Register kernel parameters

Put the kernel parameters to sabakan:
```console
$ sabactl kernel-params set "console=ttyS0 coreos.autologin=ttyS0"
```

When iPXE script is acquired, this value is passed as the kernel parameter of iPXE script.

## <a name="whatsnext" /> What's next

Learn sabakan [concepts](concepts.md), then read other specifications.

[etcd]: https://github.com/coreos/etcd
[HTTPBoot]: https://github.com/tianocore/tianocore.github.io/wiki/HTTP-Boot
[ignition]: https://coreos.com/ignition/docs/latest/
