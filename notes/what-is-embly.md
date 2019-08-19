## Embly

Embly is a safe, performant, serverless WebAssembly application runtime. It is intended to provide the benefits of a
serverless architecture without sacrificing performance. Embly functions start in <10ms, can trivially spawn
other functions, and pass messages to each other. Functions communicate by reading and writing bytes, they
can currently respond to HTTP and TCP requests, but many other protocols could be supported.

Embly functions run using Fastly's Lucet. Rust is currently the only supported language for writing
functions, but support for Typescript, C, and other languages will be added in the near future.

Embly should be considered alpha software and needs a lot of work before it's ready for productive use.

### An Embly Project

Let's quickly define some terms:

- **function**: A function is a single execution unit. It can read bytes, write bytes, spawn other functions, send/receive bytes from
  other functions, write to stdout and exit.
- **gateway**: A gateway is how a function is spawned in response to an external action. The current gateway protocols
  are TCP and HTTP. A gateway defines which function should be spawned when an HTTP or TCP request is received.
- **master**: The embly master is a single process that orchestrates all function spawning, gateways, and communication
  between functions.

An embly project is defined with an embly-project.yml file.

```yml
functions:
  - name: listener
    path: ./listener
    language: rust
gateways:
  - name: index
    type: http
    port: 8082
    function: listener
```

The project file defines functions and gateways. Each project file is intended to describe the execution structure.
These files will also include external dependencies (other projects) and details about how functions should be
hosted in various environments.

A large repo might contain one project or many projects. Functions and projects can be shared and imported. It
is intended that a project file could describe a local development setup and/or the structure of a remote deployment.

### Execution Details

The `embly` executable is run within an Embly project. The executable reads the project file and compiles
all functions from their language into WebAssembly. The resulting WebAssembly binary is then compiled into
an object file (linux and osx are supported).

After compilation the gateways are enumerated and listeners are run on the ports provided in the configuration
file. The "embly master" is then started which starts listening for function messages at the socket `/tmp/embly.sock`.

When a gateway receives a request is spawns a function. A gateway has a `uint64` address. The function is
given its own random `uint64` address. Spawning a function starts a separate process that reads the
object file for the specified function and executes it. The function starts and a handshake is performed so that the
master process knows which socket listener belongs to the function. The gateway then writes bytes to the
function.

The function can read these bytes (if it chooses) or it can write bytes back to the gateway. In the case
of the http gateway the raw bytes of the http request are sent to the function and the gateway expects a
valid http response (or it will error).

Functions can also spawn other functions. When spawning they must spawn the function by name, this name must
be registered in the Embly project file. When a spawn call is made the function creates a random `uint64`
address for the function and sends that address and the function name to the embly master. The master then
spawns the child function and handles routing messages between the functions.

Functions have access to a subset of the wasi syscalls. This allows them to generate random bytes, write to sdout
or read the time. Functions currently have no way of reading a filesystem or making external network requests.
The current idea is that "special" functions will be made available within the runtime that can be spawned for
specific tasks. The first of these functions will be an HTTP function to allow for a function to make simple
HTTP calls.
