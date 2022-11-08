# Change Log

All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]

### Changed
- Update etcd to 3.5.5 (#258)

## [2.13.1] - 2022-10-26

### Changed
- Update dependencies (#256)
    - Upgrade direct dependencies in go.mod
    - Update Golang to 1.19
    - Update GitHub Actions
    - Update etcd to 3.5.4

## [2.13.0] - 2022-04-25

### Fixed
- Clarify spec of queries in "GET machines" API (#251)
- Fix "GET machines" API for multiple serials (#251)
- **Incompatible change**: Fix "without-labels" query of "GET machines" API for clarified spec (#251)


## [2.12.0] - 2022-04-18

### Changed
- Use copy() instead of loop (#245)
- Update actions (#246)
- Update module dependencies (#247)

### Fixed
- Update incorrect link in getting_started.md (#244)

## [2.11.0] - 2022-03-23

### Changed
- Allow some state transitions. (#242)

## [2.10.0] - 2022-03-07

### Fixed
- Remove machine status metrics when the machine is removed (#240)

### Changed
- Make the options of `sabactl machines get` allow multiple values (#239) 

### Added
- Add '--without' options for `sabactl machines get`  (#239) 
- Add '-o simple' for `sabactl machines get`  (#239) 

## [2.9.2] - 2022-02-09

### Fixed
- Fix DHCP renewal is not working. (#236, #237)

## [2.9.1] - 2022-01-25

### Changed
- Increase the maximum size limit of assets to 4GiB. (#232)

## [2.9.0] - 2022-01-24

### Added
- sabakan-cryptsetup: enable no\_read\_workqueue and no\_write\_workqueue (#230)

## [2.8.0] - 2021-12-21

### Changed
- update for etcd 3.5.1 (#228)

### Fixed
- Fix some example values in the documentation. (#227)

## [2.7.1] - 2021-09-15

### Changed
- update golang to 1.17 (#223)

## [2.7.0] - 2021-08-20

### Changed
- make sabakan image upload idempotent (#220)

## [2.6.0] - 2021-05-25

### Changed
- Support etcd 3.4 as a backend store in addtion to etcd 3.3 (#217).

## [2.5.7] - 2021-05-07

### Changed
- Update etcdutil version to v1.3.6 (#215).
- Build with Go 1.16 (#211).

## [2.5.6] - 2021-02-04

### Fixed
- Fix sabakan-cryptsetup for TPM 1.2 (#208).

## [2.5.5] - 2021-02-01

### Changed
- Update dependencies
- Docker image is now built with Go 1.15 and based on Ubuntu 20.04

## [2.5.4] - 2021-01-28

### Changed
- Update go-tpm to 0.3.2 #205

## [2.5.3] - 2020-11-05

### Changed
- sabactl: Return an error when an invalid subcommand is executed (#201).

## [2.5.2] - 2020-06-08

No changes.  Only for updating Docker base image.

## [2.5.1] - 2020-04-07

### Changed
- Collect metrics synchronously instead of collecting periodically (#185).

## [2.5.0] - 2020-01-17

### Added
- Add Prometheus instrumentation (#181).

## [2.4.9] - 2019-12-26

Only cosmetic changes.

## [2.4.8] - 2019-10-28

### Changed
- Support Golang 1.13 (#173).

## [2.4.7] - 2019-09-10

No changes.  Only for updating Docker base image.

## [2.4.6] - 2019-08-23

### Changed
- Update etcd to 3.3.15 and etcdutil to 1.3.3 (#171).
- sabakan-cryptsetup: retry sabakan API calls (#171).

## [2.4.5] - 2019-08-19

### Changed
- Update etcd to 3.3.14 and etcdutil to 1.3.2 (#168).
- sabakan-cryptsetup: reformat when TPM is enabled (#169).

## [2.4.4] - 2019-07-29

### Added
- sabakan-cryptsetup: TPM 2.0 support (#164).

## [2.4.3] - 2019-07-04

### Changed
- No longer record state transition in audit (#163).

## [2.4.2] - 2019-06-03

### Changed
- Fix Dockerfile (#157).

## [2.4.1] - 2019-06-03

### Changed
- Rebuild container image to update its base image.

## [2.4.0] - 2019-04-26

### Added
- Add GraphQL API for set machine state (#156).

### Changed
- Do not push branch tag for pre-releases (#155).

## [2.3.0] - 2019-04-19

### Changed
- Refine sabakan machine state transitions (#154).

## [2.2.2] - 2019-04-18

### Changed
- Run mtest using built container image, fix bug (#153).

## [2.2.1] - 2019-04-17

### Added
- Build docker image by this repository instead of github.com/cybozu/neco-containers (#152).

### Changed
- Fix docker build job (#149, #150).
- Rebuild with Go 1.12 (#148).
- Improve mtest environment and CI (#147, #151).

## [2.2.0] - 2019-03-05

### Added
- `sabactl completion` can generate bash completion scripts (#146).

### Changed
- Transition from `retiring` to `retired` should be explicitly ordered by `sabactl machines set-state` command (#143).
- Registering disk encryption keys is prohibited for `retiring` machines in addition to `retired` ones (#143).

## [2.1.0] - 2019-02-25

### Added
- REST API to download `sabakan-cryptsetup` (#142).

### Changed
- `sabakan-cryptsetup` was rewritten (#142).

## [2.0.1] - 2019-02-19

### Added
- Ignition template can list remote files (#139).

### Changed
- Fix a critical degradation in ignition template introduced in 2.0.0 (#140).

## [2.0.0] - 2019-02-18

### Added
- Ignition templates have `version` to specify Ignition spec version for rendering (#138).
- Arithmetic functions are available in Ignition templates (#137).

### Changed
- [Semantic import versioning](https://github.com/golang/go/wiki/Modules#semantic-import-versioning) for v2 has been applied.
- REST API for Ignition templates has been revamped breaking backward-compatibility (#138).
- Go client library has been changed for new Ignition template API (#138).

## [1.2.0] - 2019-02-13

### Added
- `Machine.Info` brings NIC configuration information (#136).  
    This new information is also exposed in GraphQL and REST API.
- `ipam.json` adds new mandatory field `node-gateway-offset` (#136).  
    Existing installations continue to work thanks to automatic data conversion.

### Changed
- GraphQL data type `BMCInfoIPv4` is renamed to `NICConfig`.

### Removed
- `dhcp.json` obsoletes `gateway-offset` field (#136).  
    The field is moved to `ipam.json` as `node-gateway-offset`.

## [1.1.0] - 2019-01-29

### Added
- [ignition] `json` template function to render objects in JSON (#134).

## [1.0.1] - 2019-01-28

### Changed
- Fix a regression in ignition template introduced in #131 (#133).

## [1.0.0] - 2019-01-28

### Breaking changes
- `ipam.json` adds new mandatory field `bmc-ipv4-gateway-offset` (#132).
- Ignition template renderer sets `.` as `Machine` instead of `MachineSpec` (#132).

### Added
- `Machine` has additional information field for BMC NIC configuration (#132).

## Ancient changes

See [CHANGELOG-0](./CHANGELOG-0.md).

[Unreleased]: https://github.com/cybozu-go/sabakan/compare/v2.13.1...HEAD
[2.13.1]: https://github.com/cybozu-go/sabakan/compare/v2.13.0...v2.13.1
[2.13.0]: https://github.com/cybozu-go/sabakan/compare/v2.12.0...v2.13.0
[2.12.0]: https://github.com/cybozu-go/sabakan/compare/v2.11.0...v2.12.0
[2.11.0]: https://github.com/cybozu-go/sabakan/compare/v2.10.0...v2.11.0
[2.10.0]: https://github.com/cybozu-go/sabakan/compare/v2.9.2...v2.10.0
[2.9.2]: https://github.com/cybozu-go/sabakan/compare/v2.9.1...v2.9.2
[2.9.1]: https://github.com/cybozu-go/sabakan/compare/v2.9.0...v2.9.1
[2.9.0]: https://github.com/cybozu-go/sabakan/compare/v2.8.0...v2.9.0
[2.8.0]: https://github.com/cybozu-go/sabakan/compare/v2.7.1...v2.8.0
[2.7.1]: https://github.com/cybozu-go/sabakan/compare/v2.7.0...v2.7.1
[2.7.0]: https://github.com/cybozu-go/sabakan/compare/v2.6.0...v2.7.0
[2.6.0]: https://github.com/cybozu-go/sabakan/compare/v2.5.7...v2.6.0
[2.5.7]: https://github.com/cybozu-go/sabakan/compare/v2.5.6...v2.5.7
[2.5.6]: https://github.com/cybozu-go/sabakan/compare/v2.5.5...v2.5.6
[2.5.5]: https://github.com/cybozu-go/sabakan/compare/v2.5.4...v2.5.5
[2.5.4]: https://github.com/cybozu-go/sabakan/compare/v2.5.3...v2.5.4
[2.5.3]: https://github.com/cybozu-go/sabakan/compare/v2.5.2...v2.5.3
[2.5.2]: https://github.com/cybozu-go/sabakan/compare/v2.5.1...v2.5.2
[2.5.1]: https://github.com/cybozu-go/sabakan/compare/v2.5.0...v2.5.1
[2.5.0]: https://github.com/cybozu-go/sabakan/compare/v2.4.9...v2.5.0
[2.4.9]: https://github.com/cybozu-go/sabakan/compare/v2.4.8...v2.4.9
[2.4.8]: https://github.com/cybozu-go/sabakan/compare/v2.4.7...v2.4.8
[2.4.7]: https://github.com/cybozu-go/sabakan/compare/v2.4.6...v2.4.7
[2.4.6]: https://github.com/cybozu-go/sabakan/compare/v2.4.5...v2.4.6
[2.4.5]: https://github.com/cybozu-go/sabakan/compare/v2.4.4...v2.4.5
[2.4.4]: https://github.com/cybozu-go/sabakan/compare/v2.4.3...v2.4.4
[2.4.3]: https://github.com/cybozu-go/sabakan/compare/v2.4.2...v2.4.3
[2.4.2]: https://github.com/cybozu-go/sabakan/compare/v2.4.1...v2.4.2
[2.4.1]: https://github.com/cybozu-go/sabakan/compare/v2.4.0...v2.4.1
[2.4.0]: https://github.com/cybozu-go/sabakan/compare/v2.3.0...v2.4.0
[2.3.0]: https://github.com/cybozu-go/sabakan/compare/v2.2.2...v2.3.0
[2.2.2]: https://github.com/cybozu-go/sabakan/compare/v2.2.1...v2.2.2
[2.2.1]: https://github.com/cybozu-go/sabakan/compare/v2.2.0...v2.2.1
[2.2.0]: https://github.com/cybozu-go/sabakan/compare/v2.1.0...v2.2.0
[2.1.0]: https://github.com/cybozu-go/sabakan/compare/v2.0.1...v2.1.0
[2.0.1]: https://github.com/cybozu-go/sabakan/compare/v2.0.0...v2.0.1
[2.0.0]: https://github.com/cybozu-go/sabakan/compare/v1.2.0...v2.0.0
[1.2.0]: https://github.com/cybozu-go/sabakan/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/cybozu-go/sabakan/compare/v1.0.1...v1.1.0
[1.0.1]: https://github.com/cybozu-go/sabakan/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/cybozu-go/sabakan/compare/v0.31...v1.0.0
