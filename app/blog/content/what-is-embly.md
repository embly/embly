+++
title = "What is Embly"
date = 2020-01-09T02:03:00Z
[extra]
author = "Max McDonnell"
+++


_The project is very much a work-in-progress and things frequently break, but enough of the core patterns and functionality are implemented that it feels like a good moment to share. This blog post is also a work in progress and it will hopefully improve over time_

[Embly](https://github.com/maxmcd/embly) is a serverless webassembly[^1] framework. It allows you to run webassembly on servers with access to the basic networking and system primitives you need to get things done. Along with running webassembly Embly uses a declarative configuration for defining various webassembly services, building them, deploying them, and connecting them to useful things like dependencies, load balancers, static files and data stores.

To enable more complicated compute and networking patterns Embly roughly implements the [actor model](https://en.wikipedia.org/wiki/Actor_model), allowing webassembly [functions](#functions) to spawn other functions and send data to each other. Each function (a webassembly binary) is also safely sandboxed, allowing for multi-tenancy.

The goal is to allow a single configuration and code to define local development, build and deploy. Once deployed, Embly projects should only incur cost[^2] when used and should scale massively. You could find a shared project, run it with minimal effort and not have to worry about provisioning resources or the cost of leaving it running. 

**Overview**:
 - [Hello World](#hello-world)
 - [Components](#components)
   - [Functions](#functions)
   - [Gateways](#gateways)
   - [Datastores](#datastores)
   - [Files](#files)
   - [Syscalls](#syscalls)
 - [Philosophy/Thoughts](#philosophy-thoughts)
   - [Sharing](#sharing)
   - [Development](#development)
   - [Scale](#scale)
   - [Limitations](#limitations)
   - [Security](#security)


### Hello World

We'll start with an [hcl](https://github.com/hashicorp/hcl) configuration file in an empty directory: `embly.hcl`.

```terraform
function "hello" {
  runtime = "rust"
  path    = "./hello"
}

gateway {
  type = "http"
  port = 8080
  route "/" {
    function = "${function.hello}"
  }
}
```

In this file we've defined a function called "hello" and a gateway.

Embly functions are the core execution unit in embly. Each function has a name and some metadata so that Embly can find the function and built it into a webassembly binary. Rust is currently the only supported language and runtime in Embly so our function will be written in Rust. Our function is called "hello" and will be responding to an HTTP request. 

An Embly gateway allows us to access external systems. For this project, our gateway is a simple HTTP server that will route all requests at the root path to our function and will broadcast locally to port 8080.

Now that we have our configuration set up we'll need to create a Rust project in the `./hello` directory. Embly functions can respond to any input or output, but since we're linking our function to an HTTP gateway we'll want to use the Embly libraries HTTP primitives. We'll start by creating a `hello/Cargo.toml` and a `hello/src/main.rs` in our project directory.

```toml
[package]
name = "hello"
version = "0.0.1"

[dependencies]
embly = "0.0.4"
```

```rust
extern crate embly;
use embly::http::{Body, Request, ResponseWriter, run};
use embly::prelude::*;
use embly::Error;

fn execute(_req: Request<Body>, w: &mut ResponseWriter) -> Result<(), Error>{
    w.write_all(b"Hello World")?; // writing our hello response bytes 
    Ok(()) // if an error is returned the server will respond with an HTTP error
}

// this function is run first
fn main() -> Result<(), Error> {
    run(execute) // this is the embly::http::run function that is specific to http responses
}
```

There are a few things going on here. Our rust program has a `main()` function that Embly will run whenever it responds to a request. When an HTTP request is received our main function calls `run` and the HTTP request is passed to our `execute` function. We can then do things like ready request data, spawn other functions, access external resources, and write our response. The `ResponseWriter` can also be flushed so that a response can be returned (or streamed) during execution.

With those files in place let's run `embly run` in our directory.

```
Building function 'hello'
[hello]:    Updating crates.io index
[hello]:  Downloading crates ...
[hello]:  Downloaded embly v0.0.4
...
[hello]:  Compiling embly v0.0.4
[hello]:  Compiling hello v0.0.1 (/opt/context/hello)
[hello]:  Finished release [optimized] target(s) in 10.34s
Compiling hello.wasm to a local object file
Compilation of function 'hello' complete (5.325825ms)
Starting dev server
HTTP gateway listening on port 8080
```

The function has been built correctly and the dev server is listening at port 8080. And if we run `curl localhost:8080` we get a "Hello World" HTTP response and the following dev server log lines:

```
Started GET "/" for [::1]:59335 at 2019-10-11 21:01:04 -0400
Processing by function "function.hello"
Completed 200 OK in 13.007029ms
```



## Components

At its core, Embly is made up of a few simple components. You compile your code into [functions](#functions) that get triggered by [gateways](#gateways), call [syscalls](#syscalls), can store data in [datastores](#datastores), and can be deployed with [static files](#files). Let's discuss each of those pieces in detail.

### Functions

The core execution unit within Embly is a function. Functions are webassembly executables with access to a subset of the [wasi](https://wasi.dev) syscalls along with a small number of additional syscalls. These functions can do a few basic things:

- read and write data from other functions and gateways
- call "spawn" to run other functions or access special functionality

Each function is a wasm binary that is compiled into x86 with Fastly's [Lucet](https://github.com/fastly/lucet). 

These functions are intended to improve the experience of writing "serverless" code. Functions can start very quickly (on the order of 10ms[^3]), spawn other functions, and directly communicate. This should make it easy to write parallel workloads.

Functions can currently only communicate with functions that they spawn, but in the future functions will be able to communicate with other functions using a passed address or a message queue.

### Gateways

Gateways provide access to external resources. The only current working gateway is an HTTP server. This gateway can receive HTTP requests and run a function in response. Functions can also call external resources themselves using the `spawn` function. 

Many more gateways will be added. Websockets, TCP, SSH, WebRTC, etc. It is likely that there might be multiple implementations of each gateway. Gateways are not too difficult to implement, but each of them involves asking the question "how should we serialize the data that we're giving to the function?". The HTTP server just passes raw HTTP to the function, but in the future gateways might make heavy use of something like protobuf to simplify the language-specific support of each gateway. 

Many serverless frameworks and services implement hard timeouts on gateway actions and function runtime. Currently, Embly expects that it won't have timeouts. You'll be able to connect to a function and maintain the connection for as long as you would like. Maybe this decision will make running an Embly server operationally onerous. We'll have to see. I think timeouts are a useful pattern, but I want Embly to enable bidirectional streaming use cases that would ideally not be interrupted. 

### Datastores

Embly services (will soon) speak TCP, so if you'd like to connect embly to a traditional RDBMS, you can. However, it seemed important to try and provide datastores that map to the expected use and functionality of Embly. No fixed up-front costs, pay-per-use, minimal scale limitations, and easy multi-tenancy. Various types of key-value stores and databases fit these constraints, but for the mean-time the primary database available to an Embly function is the FoundationDB Record Layer. 

More on the Record Layer can be found here: [https://www.foundationdb.org/blog/announcing-record-layer/](https://www.foundationdb.org/blog/announcing-record-layer/)

The Record Layer does not provide a wire protocol or SDKs for various languages so Embly provides [vinyl](https://github.com/embly/vinyl) a stateless database server and protobuf wire protocol, along with an SDK for rust and Go. The client libraries are small wrappers around a grpc service so they should be easy to port to any of the languages supported by Embly.

Record Layer database can be created in milliseconds and are isolated from each other. This should allow Embly users to create and deploy data-intensive applications in seconds, without worrying about hosting complexity. 

Embly will provide first class support for managing schema migrations in Vinyl and running a database instance locally and remotely

Other datastores, blob storage, a key value cache, and message queues will be added in the future. 

### Files

Applications sometimes serve static files. Embly allows you to reference static files in a configuration that will be included in a deploy. 

Embly intends to completely define how a project can be built and deployed. Since some static files might be the result of a build script, Embly will soon add functionality to watch and run local build scripts for static files. 

### Syscalls

Embly supports the following [wasi](https://wasi.dev) syscalls (more might be added):

```
args_get
args_sizes_get
clock_time_get
environ_get
environ_sizes_get
fd_fdstat_get
fd_prestat_dir_name
fd_prestat_get
fd_write
poll_oneoff
proc_exit
random_get
```

Additionally Embly provides the following syscalls:

 - `_read` read from something (a function or a gateway)
 - `_write` write to something
 - `_spawn` spawn a function or gateway
 - `_events` get notified when something is available to read

It is very likely that something like a `close` will be added, but most resources so far manage their own lifetime. 

The intent here is to provide a very minimal implementation so that many different languages can be run on Embly with minimal effort. 

`_spawn` takes a string and is overloaded with lots of magic. `_spawn("/embly/http")` makes an HTTP request. If you need to make a db call you `_spawn("embly/vinyl/connect")` but in order to avoid the confusion of reading multiple responses from the same connection each future call is made with `_spawn("embly/vinyl/request")`. Your own functions are called with `_spawn("function/foo")` and in the future you'll be able to call imported dependencies with a similar syntax.

Maybe you think all this is terrible, maybe you think this doesn't accomplish my stated goals, I'd love to [hear your thoughts](https://github.com/maxmcd/embly/issues).

## Philosophy/Thoughts

Embly is young, so it can't do most of these things yet, but I think the overall
 structure of Embly allows for many exciting things. This section helps describe what's possible and what's next. 

### Sharing

Embly is about sharing. I have worked on [similar](https://github.com/maxmcd/gitbao) [projects](https://github.com/maxmcd/dcdn) in the past because I have been frustrated with how difficult it is for hobbyist programmers to write dynamic programs and deploy them without worrying about costs and infrastructure complexity.

Along with sharing projects I also hope that Embly can be used to share general functionality. Think of the following hypothetical user experiences:

 - "I want to try out my eCommerce idea but wouldn't like to start paying for hosting until I have users"
 - "I found this Embly package that we can use in our project. It wraps the AWS API and independently handles secret management and authentication so we don't have to deal with secrets. Since we can run it on our own Embly server we don't have to worry about a third party holding our keys"
 - "I found a functions that will index all of the code in this GitHub repository and search it"
 - "We want to move off of slack but don't want to figure out how to host our own chat service, we found a chat application on Embly and had it running and importing our Slack data in seconds"
 - "I was worried about this Chrome extension having access to my browser history and making network requests, but the backend is an Embly application that I can host myself"

Embly allows sharing of dynamic applications that store state and are addressable over the internet. This allows developers to build solutions to complicated problems that can be shared simply with technical or non-technical users. Additionally, the option of self-hosting an Embly cluster or using Embly with a hosting provider could help alleviate concerns about data privacy and misaligned data access incentives. 

Being able to easily share powerful things can be dangerous. If access to running something on Embly is readily available and your friend has a "dope DDOS application" they just wrote, then you are well positioned to wreak havoc. This is nothing new, but usually "script kiddies" have to at least follow a tutorial on how to get up and running. It's unclear how this will be solved or how exactly this problem will manifest itself.

### Development

The Embly development environment has a lot of information about what's going on. It knows what data a program accesses, what dependencies, what local files. It knows what language the project is in and how it's expected to build. It also knows what requests look like, and how they're serialized. There isn't much in the way of Embly providing a fantastic development experience. 

I remember the first time I used Ruby On Rails I was so excited about all the debugging information I got in dev logs. I think Embly can offer some of that and much more. Providing trace information for requests, allowing users to replay requests, test against production data, inject mock data, etc..

Embly could also work on the web, or on a mobile device. If build and compilation were hosted elsewhere a web or mobile editor could provide all the functionality needed to build a complete application.

### Scale

Embly should scale. Like other serverless platforms there is not much in the way of Embly being massively parallel. When a new request comes in the server needs to be able to route the request to the right function, but there's not much more from there.

Certain projects might be associated with a certain region, or use a specific datastore in a certain region. This will introduce scale constraints that need to be solved. Embly will be tasked with providing datastores and primitives that make this easy.

Part of this work is already being done with [Vinyl](#datastores). Apple uses the FoundationDB Record Layer to share schemas across the globe, but allow individual users to store their data closer to their geographic location. Embly could enable the same storage pattern. 

Data locality is also a scale and performance issue with current serverless patterns. I'm excited to explore using something similar to Google's [Slicer](https://www.usenix.org/system/files/conference/osdi16/osdi16-adya.pdf) to route requests and functions to servers that have the files or data that they need. 


### Limitations

Embly functions don't have access to threads, processes, libraries, files, sockets, certificates, good security primitives [^3] [^4], or other things that applications need. Some of these might never be supported.

Embly is slow. Currently, a process is spawned for every request, depending on the size of the wasm binary this can take many milliseconds. Making the functions faster involves navigating a balance of overhead/security/complexity, but Embly should get much faster in time.

Embly only supports languages that can compile to webassembly. 

None of your existing code will work on Embly. 

While Embly is "serverless", someone, somewhere, must run a server that is running Embly.

The state of webassembly is quite rough. Wasi isn't feature complete, language implementations are changing over time. This will surely cause issues with stability/experience in the short term.


### Security

At the moment, Embly is a very young project, is full of security issues and should not be trusted for any kind of important or secure workloads. 

Embly uses [lucet](https://github.com/fastly/lucet). So this is worth a read: [https://github.com/fastly/lucet/blob/master/SECURITY.md](https://github.com/fastly/lucet/blob/master/SECURITY.md)

Embly currently isolates each request within a process, this is inspired by some notes about v8's security model: [https://v8.dev/blog/spectre#site-isolation.](https://v8.dev/blog/spectre#site-isolation.) Maybe this is overkill, maybe it's not enough.

As mentioned in [limitations](#limitiations), wasm functions do not have access to good security primitives. These might be added as an external dependency in the future.

Nothing within Embly is encrypted, data privacy is currently ignored, this will also be improved over time.

### Notes

[^1]: There are servers, you can run Embly yourself on your own server. In describing this project to people just mentioning "serverless webassembly" seems to go a very long way in explaining what's going on, and avoiding it has more readily lead to confusion. Apologies for the bs.

[^2]: I mention cost a few times. I imagine Embly as a hosted service, similar to AWS Lambda. In this case, "cost" seems like the right framing. Someone else would be running an Embly cluster and the user would pay for what they are using. This breaks down somewhat when one is running their own cluster (and some of the scale benefits might be lost as well), but I think the framing is still helpful as you can still run Embly as a company or a group and the individuals running projects and functions would incur their own costs.

[^3]: Small webassembly functions (1-2mb binary wasm files) can currently start very quickly. Larger files are slower, and languages with runtimes are garbage collection will add further delay. There are many ways to improve these things, but for now, there is some overhead with every request. 

[^4]: [https://github.com/briansmith/ring/issues/657#issuecomment-396416821](https://github.com/briansmith/ring/issues/657#issuecomment-396416821)

[^5]: [https://github.com/CraneStation/wasmtime/issues/71#issuecomment-477330359](https://github.com/CraneStation/wasmtime/issues/71#issuecomment-477330359)
