GraphQL API
===========

Sabakan provides [GraphQL][] API at `/graphql` HTTP endpoint.

Schema
------

See [gql/schema.graphql](../gql/schema.graphql).

Playground
----------

If [sabakan](sabakan.md) starts with `-enable-playground` command-line flag,
it serves a web-based playground for GraphQL API at `/playground` HTTP endpoint.

Example
-------

Searching machines matching the these conditions:

* the machine has a label whose key is "datacenter" and value is "dc1".
* the machine's current state is _healthy_.
* the machine is not in rack 1.

can be done with a GraphQL query and variables as follows:

### Query

```
query search($having: MachineParams = null, $notHaving: MachineParams = null) {
  searchMachines(having: $having, notHaving: $notHaving) {
    spec {
      serial
      labels {
        name
        value
      }
      ipv4
      rack
    }
    status {
      state
      timestamp
      duration
    }
  }
}
```

### Variables

```json
{
    "having": {
        "labels": [
            {"name": "datacenter", "value": "dc1"}
        ],
        "states": ["HEALTHY"]
    },
    "notHaving": {
        "racks": [1]
    }
}
```

### Result

```json
{
  "data": {
    "searchMachines": [
      {
        "spec": {
          "serial": "00000004",
          "labels": [
            {
              "name": "datacenter",
              "value": "dc1"
            },
            {
              "name": "product",
              "value": "vm"
            }
          ],
          "ipv4": [
            "10.0.0.104"
          ],
          "rack": 0
        },
        "status": {
          "state": "HEALTHY",
          "timestamp": "2018-11-26T09:17:20Z",
          "duration": 21678.990289
        }
      }
    ]
  }
}
```

[GraphQL]: https://graphql.org/
