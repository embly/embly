<h2 align="center">embly</h2>

<p align="center">
  <a href="https://docs.rs/embly">
    <img src="https://docs.rs/embly/badge.svg"valign="middle"></a>
  <a href="https://crates.io/crates/embly">
    <img src="https://img.shields.io/crates/v/embly.svg"valign="middle"></a>
</p>

A serverless web application framework for collaboration and scale. 

For more background and details about what embly is read [here](https://embly.run) or [here](https://embly.run/what-is-embly)


## Hello World

 Create a new folder and add the following files and directory structure:

```
├── embly.hcl
└── hello
    ├── Cargo.toml
    └── src
        └── main.rs
```
Now add the following file contents:

`embly.hcl`:
<!-- begin embly.hcl -->
```hcl
function "hello" {
  runtime = "rust"
  path    = "./hello"
}

gateway {
  type = "http"
  port = 8765
  route "/" {
    function = "${function.hello}"
  }
}
```
<!-- end embly.hcl -->

`hello/Cargo.toml`:
<!-- begin hello/Cargo.toml -->
```toml
[package]
name = "hello"
version = "0.0.1"
edition = "2018"

[dependencies]
embly = "0.0.5"
```
<!-- end hello/Cargo.toml -->

`hello/src/main.rs`:
<!-- begin hello/src/main.rs -->
```rust
extern crate embly;
use embly::{
  http::{run_catch_error, Body, Request, ResponseWriter},
  prelude::*,
  Error,
};

async fn execute(_req: Request<Body>, mut w: ResponseWriter) -> Result<(), Error> {
  w.write_all(b"Hello World")?; // writing our hello response bytes
  Ok(()) // if an error is returned the server will respond with an HTTP error
}

// this function is run first
fn main() {
  run_catch_error(execute); // this is the embly::http::run function that is specific to http responses
}
```
<!-- end hello/src/main.rs -->

You can now run your project for local development with `embly dev`, although the fastest way to get started is with docker:

```bash
docker run -v /var/run/docker.sock:/var/run/docker.sock  -v $(pwd):/app -p 8765:8765 -it embly/embly embly dev
```

More on how to run embly in the [installation section](#Installation).


## The embly Command

```
$ embly
Usage: embly [--version] [--help] <command> [<args>]

Available commands are:
    build     Build an embly project
    bundle    Create a bundled project file
    db        Run various database maintenace tasks. 
    dev       Develop a local embly project
    run       Run a local embly project
```

## Installation

embly uses docker to download and run build images. It's recommended that you run embly from within a docker container and give it access to the docker socket. If you are in the root of an embly project you can start the dev server like so:
```bash
docker run -v /var/run/docker.sock:/var/run/docker.sock  -v $(pwd):/app -p 8765:8765 -it embly/embly embly dev
```

If you would like to run embly locally you'll need to have `cargo` and `go` installed. The following sequence of commands should work:
```bash
go get github.com/embly/embly/cmd/embly
cargo install embly-wrapper
cargo install lucetc
```


## Links

 - [homepage](https://embly/.run)
 - [what is embly?](https://embly.run/what-is-embly)
 - [rust library documentation]()
 - [example app](https://embly.run/app) and [source](/app)
 - [docker images](https://hub.docker.com/u/embly)


----

embly used to be wasabi, which was more focused on providing full operating system functionality within a
webassembly runtime. That code is [available here](https://github.com/maxmcd/wasabi-archive).
