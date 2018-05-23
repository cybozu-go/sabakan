[![CircleCI](https://circleci.com/gh/cybozu-go/sabakan.svg?style=svg)](https://circleci.com/gh/cybozu-go/sabakan)
[![GoDoc](https://godoc.org/github.com/cybozu-go/sabakan?status.svg)][godoc]
[![Go Report Card](https://goreportcard.com/badge/github.com/cybozu-go/sabakan)](https://goreportcard.com/report/github.com/cybozu-go/sabakan)

Sabakan
=======

![sabakan architecture](http://www.plantuml.com/plantuml/svg/TP3DIWCn58NtUOh32EukWtPNH4frqOMBWds1T7BxWvca93Tr8RwxYI4w3hZ9uhjVphd9AeeEaaQh0W-YtT4oEfR1OB0f2eSE7memMlHUHqOPtSt1_HmiCb2eCiZuTqTLdC4cro68B1-46lvKqwNMtWjUELpRJh-pc9lVjCFDo_buahLDh7wA7cfcSrhNFtmnvsK9vqtkBsUd_fOEOgUb3H65meWUMymIsfYUpLdwmAE_CafSJQPqcOhFcwSjRh7PxROu-82zzwBQ2xDOxYmHJqdA5_Q1luKLEvD6-mK0)
<!-- go to http://www.plantuml.com/plantuml/ and enter the above URL to edit the diagram. -->

Sabakan is an integration service to automate bare-metal server management.
It uses [etcd][] as a backend datastore for strong consistency and high availability.

**Project Status**: Initial development.

Planned features
----------------

* High availability

    Thanks to etcd, sabakan can run multiple instances while maintaining
    strong consistency.  For instance, DHCP lease information are shared
    among sabakan instances to avoid conflicts.

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

* Server life-cycle management

    Sabakan provides API to change server status for life-cycle management.

* Logs

    To track problems and life-cycle events, sabakan keeps operation logs
    within its etcd storage.

Programs
--------

This repository contains two programs:

* `sabakan`: the network service to manage servers.
* `sabactl`: CLI tool for `sabakan`.

To see their usage, run them with `-h` option.

Documentation
-------------

[docs](docs/) directory contains tutorials and specifications.

License
-------

Sabakan is licensed under MIT license.

[godoc]: https://godoc.org/github.com/cybozu-go/sabakan
[etcd]: https://coreos.com/etcd/
[HTTPBoot]: https://github.com/tianocore/tianocore.github.io/wiki/HTTP-Boot
