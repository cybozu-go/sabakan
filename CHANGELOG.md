# Change Log

All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]

## [3.1.1] - 2024-03-21

### Changed
- Update dependencies in [#281](https://github.com/cybozu-go/sabakan/pull/281)

### Fixed
- update document for sabakan TLS in [#280](https://github.com/cybozu-go/sabakan/pull/280)

## [3.1.0] - 2024-01-17

### Breaking Changes

#### Migrate image registry

We migrated the image repository of sabakan to `ghcr.io`.
From sabakan v3.1.0, please use the following image.

- https://github.com/cybozu-go/sabakan/pkgs/container/sabakan

The [quay.io/cybozu/sabakan](https://quay.io/repository/cybozu/sabakan) will not be updated in the future.

### Changed

- Migrate to ghcr.io (#276)

## [3.0.0] - 2023-11-16

### Changed
- TLS is now required for crypts API ([#270](https://github.com/cybozu-go/sabakan/pull/270))
  - see docs about [HTTPS API](https://github.com/cybozu-go/sabakan/pull/272) and [Usage](https://github.com/cybozu-go/sabakan/blob/main/docs/sabakan.md#usage)
- Bump version to v3 and update dependencies in [#273](https://github.com/cybozu-go/sabakan/pull/273)
- Update dependencies in [#266](https://github.com/cybozu-go/sabakan/pull/266)

### Fixed
- Fix to check error of etcd watch response in [#267](https://github.com/cybozu-go/sabakan/pull/267)
- Enable tests related with vTPM ([#272](https://github.com/cybozu-go/sabakan/pull/272))

## [2.13.2] - 2023-02-24

### Changed
- Update etcd to 3.5.5 ([#258](https://github.com/cybozu-go/sabakan/pull/258))
- Update dependencies in [#263](https://github.com/cybozu-go/sabakan/pull/263)
    - Upgrade direct dependencies in go.mod
    - Update testing/building environments
    - Update etcd to 3.5.7
- Fix use of math/rand in [#263](https://github.com/cybozu-go/sabakan/pull/263)
    - Go 1.20 or later is required
- Generate statically linked binaries in [#263](https://github.com/cybozu-go/sabakan/pull/263)

## [2.13.1] - 2022-10-26

### Changed
- Update dependencies ([#256](https://github.com/cybozu-go/sabakan/pull/256))
    - Upgrade direct dependencies in go.mod
    - Update Golang to 1.19
    - Update GitHub Actions
    - Update etcd to 3.5.4

## [2.13.0] - 2022-04-25

### Fixed
- Clarify spec of queries in "GET machines" API ([#251](https://github.com/cybozu-go/sabakan/pull/251))
- Fix "GET machines" API for multiple serials ([#251](https://github.com/cybozu-go/sabakan/pull/251))
- **Incompatible change**: Fix "without-labels" query of "GET machines" API for clarified spec ([#251](https://github.com/cybozu-go/sabakan/pull/251))


## [2.12.0] - 2022-04-18

### Changed
- Use copy() instead of loop ([#245](https://github.com/cybozu-go/sabakan/pull/245))
- Update actions ([#246](https://github.com/cybozu-go/sabakan/pull/246))
- Update module dependencies ([#247](https://github.com/cybozu-go/sabakan/pull/247))

### Fixed
- Update incorrect link in getting_started.md ([#244](https://github.com/cybozu-go/sabakan/pull/244))

## [2.11.0] - 2022-03-23

### Changed
- Allow some state transitions. ([#242](https://github.com/cybozu-go/sabakan/pull/242))

## [2.10.0] - 2022-03-07

### Fixed
- Remove machine status metrics when the machine is removed ([#240](https://github.com/cybozu-go/sabakan/pull/240))

### Changed
- Make the options of `sabactl machines get` allow multiple values ([#239](https://github.com/cybozu-go/sabakan/pull/239)) 

### Added
- Add '--without' options for `sabactl machines get`  ([#239](https://github.com/cybozu-go/sabakan/pull/239)) 
- Add '-o simple' for `sabactl machines get`  ([#239](https://github.com/cybozu-go/sabakan/pull/239)) 

## [2.9.2] - 2022-02-09

### Fixed
- Fix DHCP renewal is not working. ([#236](https://github.com/cybozu-go/sabakan/pull/236), [#237](https://github.com/cybozu-go/sabakan/pull/237))

## [2.9.1] - 2022-01-25

### Changed
- Increase the maximum size limit of assets to 4GiB. ([#232](https://github.com/cybozu-go/sabakan/pull/232))

## [2.9.0] - 2022-01-24

### Added
- sabakan-cryptsetup: enable no\_read\_workqueue and no\_write\_workqueue ([#230](https://github.com/cybozu-go/sabakan/pull/230))

## [2.8.0] - 2021-12-21

### Changed
- update for etcd 3.5.1 ([#228](https://github.com/cybozu-go/sabakan/pull/228))

### Fixed
- Fix some example values in the documentation. ([#227](https://github.com/cybozu-go/sabakan/pull/227))

## [2.7.1] - 2021-09-15

### Changed
- update golang to 1.17 ([#223](https://github.com/cybozu-go/sabakan/pull/223))

## [2.7.0] - 2021-08-20

### Changed
- make sabakan image upload idempotent ([#220](https://github.com/cybozu-go/sabakan/pull/220))

## [2.6.0] - 2021-05-25

### Changed
- Support etcd 3.4 as a backend store in addtion to etcd 3.3 ([#217](https://github.com/cybozu-go/sabakan/pull/217)).

## [2.5.7] - 2021-05-07

### Changed
- Update etcdutil version to v1.3.6 ([#215](https://github.com/cybozu-go/sabakan/pull/215)).
- Build with Go 1.16 ([#211](https://github.com/cybozu-go/sabakan/pull/211)).

## [2.5.6] - 2021-02-04

### Fixed
- Fix sabakan-cryptsetup for TPM 1.2 ([#208](https://github.com/cybozu-go/sabakan/pull/208)).

## [2.5.5] - 2021-02-01

### Changed
- Update dependencies
- Docker image is now built with Go 1.15 and based on Ubuntu 20.04

## [2.5.4] - 2021-01-28

### Changed
- Update go-tpm to 0.3.2 [#205](https://github.com/cybozu-go/sabakan/pull/205)

## [2.5.3] - 2020-11-05

### Changed
- sabactl: Return an error when an invalid subcommand is executed ([#201](https://github.com/cybozu-go/sabakan/pull/201)).

## [2.5.2] - 2020-06-08

No changes.  Only for updating Docker base image.

## [2.5.1] - 2020-04-07

### Changed
- Collect metrics synchronously instead of collecting periodically ([#185](https://github.com/cybozu-go/sabakan/pull/185)).

## [2.5.0] - 2020-01-17

### Added
- Add Prometheus instrumentation ([#181](https://github.com/cybozu-go/sabakan/pull/181)).

## [2.4.9] - 2019-12-26

Only cosmetic changes.

## [2.4.8] - 2019-10-28

### Changed
- Support Golang 1.13 ([#173](https://github.com/cybozu-go/sabakan/pull/173)).

## [2.4.7] - 2019-09-10

No changes.  Only for updating Docker base image.

## [2.4.6] - 2019-08-23

### Changed
- Update etcd to 3.3.15 and etcdutil to 1.3.3 ([#171](https://github.com/cybozu-go/sabakan/pull/171)).
- sabakan-cryptsetup: retry sabakan API calls ([#171](https://github.com/cybozu-go/sabakan/pull/171)).

## [2.4.5] - 2019-08-19

### Changed
- Update etcd to 3.3.14 and etcdutil to 1.3.2 ([#168](https://github.com/cybozu-go/sabakan/pull/168)).
- sabakan-cryptsetup: reformat when TPM is enabled ([#169](https://github.com/cybozu-go/sabakan/pull/169)).

## [2.4.4] - 2019-07-29

### Added
- sabakan-cryptsetup: TPM 2.0 support ([#164](https://github.com/cybozu-go/sabakan/pull/164)).

## [2.4.3] - 2019-07-04

### Changed
- No longer record state transition in audit ([#163](https://github.com/cybozu-go/sabakan/pull/163)).

## [2.4.2] - 2019-06-03

### Changed
- Fix Dockerfile ([#157](https://github.com/cybozu-go/sabakan/pull/157)).

## [2.4.1] - 2019-06-03

### Changed
- Rebuild container image to update its base image.

## [2.4.0] - 2019-04-26

### Added
- Add GraphQL API for set machine state ([#156](https://github.com/cybozu-go/sabakan/pull/156)).

### Changed
- Do not push branch tag for pre-releases ([#155](https://github.com/cybozu-go/sabakan/pull/155)).

## [2.3.0] - 2019-04-19

### Changed
- Refine sabakan machine state transitions ([#154](https://github.com/cybozu-go/sabakan/pull/154)).

## [2.2.2] - 2019-04-18

### Changed
- Run mtest using built container image, fix bug ([#153](https://github.com/cybozu-go/sabakan/pull/153)).

## [2.2.1] - 2019-04-17

### Added
- Build docker image by this repository instead of github.com/cybozu/neco-containers ([#152](https://github.com/cybozu-go/sabakan/pull/152)).

### Changed
- Fix docker build job ([#149](https://github.com/cybozu-go/sabakan/pull/149), [#150](https://github.com/cybozu-go/sabakan/pull/150)).
- Rebuild with Go 1.12 ([#148](https://github.com/cybozu-go/sabakan/pull/148)).
- Improve mtest environment and CI ([#147](https://github.com/cybozu-go/sabakan/pull/147), [#151](https://github.com/cybozu-go/sabakan/pull/151)).

## [2.2.0] - 2019-03-05

### Added
- `sabactl completion` can generate bash completion scripts ([#146](https://github.com/cybozu-go/sabakan/pull/146)).

### Changed
- Transition from `retiring` to `retired` should be explicitly ordered by `sabactl machines set-state` command ([#143](https://github.com/cybozu-go/sabakan/pull/143)).
- Registering disk encryption keys is prohibited for `retiring` machines in addition to `retired` ones ([#143](https://github.com/cybozu-go/sabakan/pull/143)).

## [2.1.0] - 2019-02-25

### Added
- REST API to download `sabakan-cryptsetup` ([#142](https://github.com/cybozu-go/sabakan/pull/142)).

### Changed
- `sabakan-cryptsetup` was rewritten ([#142](https://github.com/cybozu-go/sabakan/pull/142)).

## [2.0.1] - 2019-02-19

### Added
- Ignition template can list remote files ([#139](https://github.com/cybozu-go/sabakan/pull/139)).

### Changed
- Fix a critical degradation in ignition template introduced in 2.0.0 ([#140](https://github.com/cybozu-go/sabakan/pull/140)).

## [2.0.0] - 2019-02-18

### Added
- Ignition templates have `version` to specify Ignition spec version for rendering ([#138](https://github.com/cybozu-go/sabakan/pull/138)).
- Arithmetic functions are available in Ignition templates ([#137](https://github.com/cybozu-go/sabakan/pull/137)).

### Changed
- [Semantic import versioning](https://github.com/golang/go/wiki/Modules#semantic-import-versioning) for v2 has been applied.
- REST API for Ignition templates has been revamped breaking backward-compatibility ([#138](https://github.com/cybozu-go/sabakan/pull/138)).
- Go client library has been changed for new Ignition template API ([#138](https://github.com/cybozu-go/sabakan/pull/138)).

## [1.2.0] - 2019-02-13

### Added
- `Machine.Info` brings NIC configuration information ([#136](https://github.com/cybozu-go/sabakan/pull/136)).
    This new information is also exposed in GraphQL and REST API.
- `ipam.json` adds new mandatory field `node-gateway-offset` ([#136](https://github.com/cybozu-go/sabakan/pull/136)).
    Existing installations continue to work thanks to automatic data conversion.

### Changed
- GraphQL data type `BMCInfoIPv4` is renamed to `NICConfig`.

### Removed
- `dhcp.json` obsoletes `gateway-offset` field ([#136](https://github.com/cybozu-go/sabakan/pull/136)).
    The field is moved to `ipam.json` as `node-gateway-offset`.

## [1.1.0] - 2019-01-29

### Added
- [ignition] `json` template function to render objects in JSON ([#134](https://github.com/cybozu-go/sabakan/pull/134)).

## [1.0.1] - 2019-01-28

### Changed
- Fix a regression in ignition template introduced in [#131](https://github.com/cybozu-go/sabakan/pull/131) ([#133](https://github.com/cybozu-go/sabakan/pull/133)).

## [1.0.0] - 2019-01-28

### Breaking changes
- `ipam.json` adds new mandatory field `bmc-ipv4-gateway-offset` ([#132](https://github.com/cybozu-go/sabakan/pull/132)).
- Ignition template renderer sets `.` as `Machine` instead of `MachineSpec` ([#132](https://github.com/cybozu-go/sabakan/pull/132)).

### Added
- `Machine` has additional information field for BMC NIC configuration ([#132](https://github.com/cybozu-go/sabakan/pull/132)).

## Ancient changes

See [CHANGELOG-0](./CHANGELOG-0.md).

[Unreleased]: https://github.com/cybozu-go/sabakan/compare/v3.1.1...HEAD
[3.1.1]: https://github.com/cybozu-go/sabakan/compare/v3.1.0...v3.1.1
[3.1.0]: https://github.com/cybozu-go/sabakan/compare/v3.0.0...v3.1.0
[3.0.0]: https://github.com/cybozu-go/sabakan/compare/v2.13.2...v3.0.0
[2.13.2]: https://github.com/cybozu-go/sabakan/compare/v2.13.1...v2.13.2
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
