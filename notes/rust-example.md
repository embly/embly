# Embly

Embly is a lightweight application runtime. It runs small isolated programs that we call "functions". Functions can do a handful of things:

- Receive bytes
- Send bytes
- Spawn a new function

Let's go through a few code examples of how to use functions. Functions are compiled to
webassembly so for our examples we'll use Rust.

### Receive Bytes

When a function begins execution it can optionally read in any bytes that it might have
been sent. Maybe there are bytes ready on startup, maybe it'll receive them later.

Over time, a function can receive multiple messages. Maybe parts of a request body or
various incremental updates. Each separate message will be separated by an `io::EOF`
error.

```rust
use embly::Comm

fn entrypoint(comm: Comm) -> io::Result<()> {
    let mut buffer = Vec::new();
    // Comm implements std::io::Read
    comm.read_to_end(&mut buffer)?;

    // a little while later you might get another message
    comm.read_to_end(&mut buffer)?;

}
```

### Write Bytes

Writes can be written back. A function is always executed by something. This could be a
command line call, a load balancer or another function. Writing to a comm will send
those bytes back to the function runner.

```rust
use embly::Comm

fn entrypoint(comm: Comm) -> io::Result<()> {
    // you can call write_all to send one message
    comm.write_all("Hello World".as_bytes())?


    // Or you can make multiple calls with write if you want to construct a
    // message and then flush the response
    comm.write(b"Hello")?
    comm.write(b"World")?
    comm.flush()?
}
```

### Spawn a function

You can spawn any function by name. You'll receive a handler from the function that can be used
to read or write data.

```rust
use embly::Comm
use embly::Function

fn entrypoint(comm: Comm) -> io::Result<()> {
    let mut foo_comm = Function::spawn("github.com/maxmcd/foo", "Hello".as_bytes())?;
    // can send an empty vec if your spark doesn't expect
    // to receive any bytes
    let mut foo_comm = Function::spawn("github.com/maxmcd/foo", vec![])?;
    foo_comm.write_all(" foo".as_bytes())?;


    // get a response back from  foo
    let mut buffer = Vec::new();
    foo_comm.read_to_end(&mut buffer)?;
}
```
