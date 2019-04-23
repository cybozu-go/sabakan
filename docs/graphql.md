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

Examples
--------

* Query
  - [machine](#example-machine)
  - [searchMachines](#example-searchmachines)
* Mutation
  - [setMachineState](#example-setmachinestate)

Example: `machine`
------------------

Query:

```graphql
query get($serial: ID!) {
  machine(serial: $serial) {
    spec {
      serial
      ipv4
    }
    status {
      state
    }
  }
}
```

Variables:

```json
{
  "serial": "00000004"
}
```

### Successful response

Get a machine with a given serial. It can be done with a GraphQL query and variables as follows:

Result:

```json
{
  "data": {
    "machine": {
      "spec": {
        "serial": "00000004",
        "ipv4": [
          "10.0.0.104"
        ]
      },
      "status": {
        "state": "UNHEALTHY"
      }
    }
  }
}
```

### Failure response

```json
{
  "errors": [
    {
      "message": "not found",
      "path": [
        "machine"
      ]
    }
  ],
  "data": null
}
```

Example: `searchMachines`
-------------------------

Query:

```graphql
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

Variables:

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

### Successful response

Searching machines matching these conditions:

* the machine has a label whose key is "datacenter" and value is "dc1".
* the machine's current state is _healthy_.
* the machine is not in rack 1.

can be done with a GraphQL query and variables as follows:

Result:

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

### Failure responses

- No such machines found.

Result:

```json
{
  "data": {
    "searchMachines": []
  }
}
```

Example: `setMachineState`
--------------------------

Query:

```graphql
mutation {
  setMachineState(serial: "00000004", state: UNHEALTHY) {
    state
  }
}
```

### Successful response

Set machine state _unhealthy_.

* the machine's current state is _healthy_.
* the machine's serial is _00000004_.

can be done with a GraphQL query and variables as follows:

Result:

```json
{
   "data": {
     "setMachineState": {
       "state": "UNHEALTHY"
     }
   }
 }
```

### Failure responses

- Invalid state value.

```json
{
  "errors": [
    {
      "message": "invalid state: RETIRE",
      "path": [
        "setMachineState"
      ],
      "extensions": {
        "type": "INVALID_STATE_NAME"
      }
    }
  ],
  "data": null
}
```

- Transitioning a retiring server to retired that still has disk encryption keys.

```json
{
  "errors": [
    {
      "message": "encryption key exists",
      "path": [
        "setMachineState"
      ],
      "extensions": {
        "serial": "00000004",
        "type": "ENCRYPTION_KEY_EXISTS"
      }
    }
  ],
  "data": null
}
```

- No specified machine found.

```json
{
  "errors": [
    {
      "message": "not found",
      "path": [
        "setMachineState"
      ],
      "extensions": {
        "serial": "00000007",
        "type": "MACHINE_NOT_FOUND"
      }
    }
  ],
  "data": null
}
```

- Invalid state transition.

Result:

```json
{
  "errors": [
    {
      "message": "transition from [ healthy ] to [ uninitialized ] is forbidden",
      "path": [
        "setMachineState"
      ],
      "extensions": {
        "serial": "00000004",
        "type": "INVALID_STATE_TRANSITION"
      }
    }
  ],
  "data": null
}
```

[GraphQL]: https://graphql.org/
