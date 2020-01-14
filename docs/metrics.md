Metrics
=======

Sabakan exposes the following metrics with the Prometheus format. The listen address and update interval can be configured by the CLI flags (see [here](sabakan.md#Usage)). All these metrics are prefixed with `sabakan_`

| Name               | Description                                                          | Type    | Labels                                                   |
| ------------------ | -------------------------------------------------------------------- | ------- | -------------------------------------------------------- |
| machine_status     | The machine status (see [Machine States](lifecycle.m#Machine-States) | Gauge   | status, address, serial, rack, index, role, machine_type |
| api_request_count  | The request counts of API call.                                      | Counter | code, path, verb                                         |
| assets_bytes_total | The total byte size of assets.                                       | Gauge   | (none)                                                   |
| assets_items_total | The total item numbers of assets.                                    | Gauge   | (none)                                                   |
| images_bytes_total | The total byte size of images.                                       | Gauge   | (none)                                                   |
| images_items_total | The total item numbers of images.                                    | Gauge   | (none)                                                   |

Note that sabakan also exposes the metrics provided by the Prometheus client library which located under `go` and `process` namespaces.
