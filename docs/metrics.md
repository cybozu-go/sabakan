Metrics
=======

Sabakan exposes the following metrics with the Prometheus format. The listen address can be configured by the CLI flag (see [here](sabakan.md#Usage)). All these metrics are prefixed with `sabakan_`

| Name               | Description                                                            | Type    | Labels                                                |
| ------------------ | ---------------------------------------------------------------------- | ------- | ----------------------------------------------------- |
| machine_status     | The machine status (see [Machine States](lifecycle.md#Machine-States)) | Gauge   | status, address, serial, rack, role, machine_type (*) |
| api_request_count  | The request counts of API call.                                        | Counter | code, path, verb                                      |
| assets_bytes_total | The total byte size of assets.                                         | Gauge   |                                                       |
| assets_items_total | The total item numbers of assets.                                      | Gauge   |                                                       |
| images_bytes_total | The total byte size of images.                                         | Gauge   |                                                       |
| images_items_total | The total item numbers of images.                                      | Gauge   |                                                       |

Note that sabakan also exposes the metrics provided by the Prometheus client library which located under `go` and `process` namespaces.

(*) "machine_type" is derived from [the user-defined `labels`](machine.md#machinespec-struct) with the key of `machine-type`.
