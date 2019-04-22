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
  - setMachineState([Valid example](#example-setmachinestate), [Invalid example](#example-setmachinestateinvalid-state-transitions))

Example: `machine`
------------------

Get a machine specified serial:

can be done with a GraphQL query and variables as follows:

### Query

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

### Variables

```json
{
  "serial": "00000004"
}
```

### Result

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

Example: `searchMachines`
-------------------------

Searching machines matching the these conditions:

* the machine has a label whose key is "datacenter" and value is "dc1".
* the machine's current state is _healthy_.
* the machine is not in rack 1.

can be done with a GraphQL query and variables as follows:

### Query

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

Example: `setMachineState`
--------------------------

Set machine state _unhealthy_:

* the machine's current state is _healthy_.
* the machine's serial is _00000004_.

can be done with a GraphQL query and variables as follows:

### Query

```graphql
mutation setState($serial: ID!, $state: MachineState!) {
  setMachineState(serial: $serial, state: $state) {
    state
  }
}
```

### Variables

```json
{
  "serial": "00000004",
  "state": "UNHEALTHY"
}
```

### Result

```json
{
   "data": {
     "setMachineState": {
       "state": "UNHEALTHY"
     }
   }
 }
```

Example: `setMachineState`(Invalid state transitions)
-----------------------------------------------------

Transition from _healthy_ to _uninitialized_ is not permitted.
`setMachineState` returns error including _extensions_ when it receives a request for an invalid state transition.

### Query

```graphql
mutation setState($serial: ID!, $state: MachineState!) {
  setMachineState(serial: $serial, state: $state) {
    state
  }
}
```

### Variables

```json
{
  "serial": "00000004",
  "state": "UNINITIALIZED"
}
```

### Result

```json
{
  "errors": [
    {
      "message": "transition from [ healthy ] to [ uninitialized ] is forbidden",
      "path": [
        "setMachineState"
      ],
      "extensions": {
        "from": "healthy",
        "to": "uninitialized"
      }
    }
  ],
  "data": null
}
```

[GraphQL]: https://graphql.org/
