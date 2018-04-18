[![CircleCI](https://circleci.com/gh/cybozu-go/sabakan.svg?style=svg)](https://circleci.com/gh/cybozu-go/sabakan)
[![GoDoc](https://godoc.org/github.com/cybozu-go/sabakan?status.svg)][godoc]
[![Go Report Card](https://goreportcard.com/badge/github.com/cybozu-go/sabakan)](https://goreportcard.com/report/github.com/cybozu-go/sabakan)

# Sabakan

Sabakan is the Bare-metal management system to organize hardware information using etcd as a backend distributed database.

## Features

* Server inventory

Sabakan provides helpful information depending on an organization's requirements.

* Disc encryption inventory

Sabakan managed servers requires disk encryption for variable data. When a server initializes Operating system, It formats target storage with encryption key. The key is stored in the Sabakan. 

## Plan

* HTTP server

All Sabakan managed servers are started up by HTTP boot. HTTP server provides kernel and initramfs to boot them.

* DHCP server

Sabakan(or another software) distributes IP address to servers for HTTP boot. Administrators have to define machines in Sabakan before servers start. DHCP server get defined IP addresses in Sabakan, then set them as static DHCP IP addresses.

## Usage

This project provides two commands, `sabakan` and `sabactl`.
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

- [etcd](https://github.com/coreos/etcd)

### Install sabakan and sabactl

Install `sabakan` and `sabactl`:

```console
$ go get -u github.com/cybozu-go/sabakan/cmd/sabakan
$ go get -u github.com/cybozu-go/sabakan/cmd/sabactl
```

Specification
-------------

See [SPEC](SPEC.md).

License
-------

MIT

[godoc]: https://godoc.org/github.com/cybozu-go/sabakan
