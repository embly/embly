# Embly Configuration

## Function

Functions have a name. They have a path and context so that it's clear how and where to build.
They have a language that maps to a supported compiler.

`service_definition` is very speculative, but it would allow one to pass a protobuf
definition and have that protobuf service map to the expected input/output of this
function (see [schemas](#schemas))

```terraform
function "encoder" {
  path     = "./examples/mjpeg/encoder"
  context  = "../../.."
  language = "rust"

  service_definition = "${schema.proto.EncoderService}"
}
```

## Gateway

The name might be wrong here, gateways could be in and out, but here we use them
to refer to a listening/broadcasting point. If an embly project file maps to a
running application then it would make sense to allow one to define routes and
rules on a single broadcasting IP. To accomplish this, we could allow a gateway
to define sub-routes.

A gateway can also point to a single function, or it could be passed a preprocessor
that would wrap all requests before sending them on.

```terraform
gateway {
  type = "http"
  port = 8080
  route "/" {
    function = "${function.encoder}"
  }
  route "/foo" {
    function = "${function.encoder}"
  }
}
```

## Dependencies

This section is very unclear. The idea here is that you would define all dependencies
used by this project. Those files/functions would be pulled in at build/deploy and
functions would be disallowed from calling to anything else.

Spawning is currently done with strings, so we could call the full dependency
string or allow the value to be mapped.

Dependencies also might be in a registry, a git repo, or maybe the local filesystem.

Lots to think about.

```terraform
dependencies = [
  "embly/http",
  "embly/tcp",
]
```

## Files

References to files that will be uploaded with the deploy. These files could be serve
with http, or they could be accessible by the functions themselves.

Might be safe for now to define file serving with the gateway.

```terraform
files "assets" {
  path = "./"
}
```

## Database

There are several persistence stores that are going to be made available to functions. The idea
so far is that there will be:

- global slow KV store
- local cache (redis/memcache)
- foundationdb record store

It is very helpful to define these things in this project file for a few reasons:

- If the project file is deployed, the database might be the only thing that has strict
  locality restrictions. Therefore a project without a db can be run at any edge server.
- This could be a place to define schemas, migration logic is still complicated, but
  would mean that in the short term it's very easy to set the data structure
- The dev environment can be easily spun up with the necessary database

```terraform
database "main" {
  type = "kv"
}
```

## Schema

This is also very rough. General idea is that if the project can define schemas and
structures that are passed between functions then we can generate libraries for each
language to easily spawn and communicate with other functions

Later support for something like flatbuffers might also get us closer to zero
copy performance between certain functions.

```terraform
schema "proto" {
  path = "./"
  type = "proto"
}
```
