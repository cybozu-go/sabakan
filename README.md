[![CircleCI](https://circleci.com/gh/cybozu-go/sabakan.svg?style=svg)](https://circleci.com/gh/cybozu-go/sabakan)
[![GoDoc](https://godoc.org/github.com/cybozu-go/sabakan?status.svg)][godoc]
[![Go Report Card](https://goreportcard.com/badge/github.com/cybozu-go/sabakan)](https://goreportcard.com/report/github.com/cybozu-go/sabakan)

Sabakan
=======

![sabakan architecture](http://www.plantuml.com/plantuml/svg/TP3DIWCn58NtUOh32EukWtPNH4frqOMBWds1T7BxWvca93Tr8RwxYI4w3hZ9uhjVphd9AeeEaaQh0W-YtT4oEfR1OB0f2eSE7memMlHUHqOPtSt1_HmiCb2eCiZuTqTLdC4cro68B1-46lvKqwNMtWjUELpRJh-pc9lVjCFDo_buahLDh7wA7cfcSrhNFtmnvsK9vqtkBsUd_fOEOgUb3H65meWUMymIsfYUpLdwmAE_CafSJQPqcOhFcwSjRh7PxROu-82zzwBQ2xDOxYmHJqdA5_Q1luKLEvD6-mK0)
<!-- go to http://www.plantuml.com/plantuml/ and enter the above URL to edit the diagram. -->

Sabakan is an integration service to automate bare-metal server management.
It uses [etcd][] as a backend datastore for strong consistency and high availability.

Features
--------

* Server inventory

    Servers in a data center can be registered with sabakan's inventory.
    In addition, sabakan assigns IP addresses automatically to servers.

* DHCP service

    Sabakan provides DHCP server service for network boot of servers.
    The DHCP implementation is primarily for [UEFI HTTP Boot][HTTPBoot].

* HTTP service

    Sabakan provides HTTP service for network boot of servers.

* Encryption key store

    Server storages often need to be encrypted.
    Sabakan provides REST API to store and retrieve encryption keys.

Usage
-----

This repository contains two programs: `sabakan` and `sabactl`.
`sabakan` is a daemon to manage servers.
`sabactl` is a helper command to control `sabakan`.

### sabakan

`sabakan` requires etcd endpoints to control data in the key-value store.

```console
$ sabakan \
   --http 0.0.0.0:8888 \
   --etcd-servers http://etcd1.example.com:2379,http://etcd2.example.com:2379 \
   --etcd-prefix sabakan

Options:
   --http
      <Listen IP>:<Port number> (default "0.0.0.0:8888")
   --etcd-servers
      URLs of the backend etcd (default "http://localhost:2379")
   --etcd-prefix
      etcd prefix
```

### sabactl command

`sabactl` reads a json file in command-line arguments, then create Sabakan config and server inventory, or update and delete them.


```console
$ sabactl <flags> <subcommand> <subcommand args>

Subcommands:
   commands         list all command names
   flags            describe all known top-level flags
   help             describe subcommands and their syntax
   remote-config    Configure a sabakan server.

$ sabactl flags

Options:
   -server
      <Listen IP>:<Port number> (default "http://localhost:8888")

$ sabactl remote-config

Subcommands:
   get              Get a sabakan server config.
   set              Set a sabakan server config.
```

## Getting started

### Prerequisites

- [etcd3][]

### Install sabakan and sabactl

Install `sabakan` and `sabactl`:

```console
$ go get -u github.com/cybozu-go/sabakan/cmd/sabakan
$ go get -u github.com/cybozu-go/sabakan/cmd/sabactl
```

### Debugging with a DHCP client

`make debug` launch a DHCP client note for the debugging:

```console
$ make debug
```

It launch etcd on rkt, create a bridge, and launch VM by QEMU.  Use `make connect` to connect the VM.

```console
$ make connect
```

Specification
-------------

See [SPEC](SPEC.md).

License
-------

Sabakan is licensed under MIT license.

The source code contains a [Netboot](https://github.com/google/netboot) which is licensed under Apache License 2.0.

[godoc]: https://godoc.org/github.com/cybozu-go/sabakan
[etcd]: https://coreos.com/etcd/
[HTTPBoot]: https://github.com/tianocore/tianocore.github.io/wiki/HTTP-Boot
