//! Embly is a serverless webassembly runtime. It runs small isolated functions.
//! Functions can do a handful of things:
//!
//! - Receive bytes
//! - Send bytes
//! - Spawn a new function
//!
//! This library is used to access embly functionality from within a program
//! being run by Embly.
//!

#![deny(
    missing_docs,
    trivial_numeric_casts,
    unstable_features,
    unused_extern_crates,
    unused_features
)]
#![warn(unused_import_braces, unused_parens)]
#![cfg_attr(feature = "clippy", plugin(clippy(conf_file = "../../clippy.toml")))]
#![cfg_attr(
    feature = "cargo-clippy",
    allow(clippy::new_without_default, clippy::new_without_default)
)]
#![cfg_attr(
    feature = "cargo-clippy",
    warn(
        clippy::float_arithmetic,
        clippy::mut_mut,
        clippy::nonminimal_bool,
        clippy::option_map_unwrap_or,
        clippy::option_map_unwrap_or_else,
        clippy::unicode_not_nfc,
        clippy::use_self
    )
)]

pub use failure::Error;
pub mod error;
pub mod http;
mod http_proto;
pub mod kv;
mod proto;
mod task;

use crate::prelude::*;
use std::{
    future::Future,
    pin::Pin,
    task::{Context, Poll},
    {io, time},
};

pub mod prelude {
    //! A "prelude" for crates using the `embly` crate
    //!
    //! imports io::Read and io::Write
    //! ```
    //! pub use std::io::Read as _;
    //! pub use std::io::Write as _;
    //! ```
    pub use std::io::Read as _;
    pub use std::io::Write as _;
}

use std::sync::Mutex;
#[macro_use]
extern crate lazy_static;

use std::collections::btree_set::BTreeSet;
lazy_static! {
    static ref EVENT_REGISTRY: Mutex<BTreeSet<i32>> = { Mutex::new(BTreeSet::new()) };
    // static ref EVENT_WAKER_REGISTRY: Mutex<HashMap<i32, Waker>> = { Mutex::new(HashMap::new()) };
}

/// Connections that handle communication between functions or gateways
///
/// ## Receive Bytes
///
/// When a function begins execution it can optionally read in any bytes that it might have
/// been sent. Maybe there are bytes ready on startup, maybe it'll receive them later.
///
///
/// ```rust
/// use embly::{Conn, Error};
/// use embly::prelude::*;
///
/// fn entrypoint(mut conn: Conn) -> Result<(), Error> {
///     let mut buffer = Vec::new();
///     // Conn implements std::io::Read
///     conn.wait()?;
///     conn.read_to_end(&mut buffer)?;
///     
///     // a little while later you might get another message
///     conn.wait()?;
///     conn.read_to_end(&mut buffer)?;
///     return Ok(())
/// }
/// ```
///
/// ## Write Bytes
///
/// Bytes can be written back. A function is always executed by something. This could be a
/// command line call, a load balancer or another function. Writing to a connection will send
/// those bytes back to the function runner.
///
/// ```rust
/// use embly::Conn;
/// use embly::prelude::*;
/// use std::io;
///
/// fn entrypoint(mut conn: Conn) -> io::Result<()> {
///     // you can call write_all to send one message
///     conn.write_all("Hello World".as_bytes())?;
///
///
///     // Or you can make multiple calls with write if you want to construct a
///     // message and then flush the response
///     conn.write(b"Hello")?;
///     conn.write(b"World")?;
///     conn.flush()?;
///     return Ok(())
/// }
/// ```
///
///
///

#[derive(Debug)]
pub struct Conn {
    id: i32,
    polled: bool,
}

impl Conn {
    fn new(id: i32) -> Self {
        Self { id, polled: false }
    }

    /// Read any bytes available on the connection and return them.
    pub fn bytes(&mut self) -> Result<Vec<u8>, Error> {
        let mut buffer = Vec::new();
        self.read_to_end(&mut buffer)?;
        Ok(buffer)
    }
    /// Read bytes available on the connection and cast them to a string
    pub fn string(&mut self) -> Result<String, Error> {
        let mut buffer = String::new();
        self.read_to_string(&mut buffer)?;
        Ok(buffer)
    }
    /// Wait for bytes to be available on the connection
    /// ```
    /// use embly::{
    ///     Error,
    ///     prelude::*,
    /// };
    /// fn run(conn: embly::Conn) -> Result<(), Error> {
    ///     conn.wait()
    /// }
    /// ```
    ///
    /// Conn implements `Future` so better to await instead:
    /// ```
    /// use embly::{
    ///     Error,
    ///     prelude::*,
    /// };
    ///
    /// async fn run(conn: embly::Conn) -> Result<(), Error> {
    ///     conn.await
    /// }
    ///
    /// ```
    pub fn wait(&self) -> Result<(), Error> {
        wait_id(self.id)
    }
}

impl Copy for Conn {}
impl Clone for Conn {
    fn clone(&self) -> Self {
        Self {
            id: self.id,
            polled: self.polled,
        }
    }
}

/// Spawn a Function
///
/// ```
/// use embly::{Conn, spawn_function};
/// use embly::prelude::*;
/// use failure::Error;
///
/// fn entrypoint(conn: Conn) -> Result<(), Error> {
///     let mut foo = spawn_function("github.com/maxmcd/foo")?;
///     foo.write_all("Hello".as_bytes())?;
///
///     // get a response back from  foo
///     let mut buffer = Vec::new();
///     foo.read_to_end(&mut buffer)?;
///     Ok(())
/// }
///
/// ```
///
pub fn spawn_function(name: &str) -> Result<Conn, Error> {
    Ok(Conn::new(spawn(name)?))
}

/// spawn, but then immediately send bytes along afterward
pub fn spawn_and_send(name: &str, payload: &[u8]) -> Result<Conn, Error> {
    let mut conn = spawn_function(name)?;
    conn.write(&payload)?;
    Ok(conn)
}

fn process_event_ids(ids: Vec<i32>) {
    let mut er = EVENT_REGISTRY.lock().unwrap();
    // let mut ewr = EVENT_WAKER_REGISTRY.lock().unwrap();
    for id in ids {
        // if let Some(waker) = ewr.remove(&id) {
        //     waker.wake()
        // }
        er.insert(id);
    }
}

fn has_id(id: i32) -> bool {
    let er = EVENT_REGISTRY.lock().unwrap();
    er.contains(&id)
}

// only used by wasm
#[allow(dead_code)]
fn remove_id(id: i32) {
    let mut er = EVENT_REGISTRY.lock().unwrap();
    er.remove(&id);
}

impl Future for Conn {
    type Output = Result<(), Error>;

    fn poll(mut self: Pin<&mut Self>, _cx: &mut Context) -> Poll<Self::Output> {
        // Check for available events if we've never polled before, block
        // indefinitely if we have
        let timeout = if self.polled {
            None
        } else {
            self.polled = true;
            Some(time::Duration::new(0, 0))
        };
        let ids = events(timeout).expect("how do we handle this error");
        process_event_ids(ids);
        if has_id(self.id) {
            self.polled = false;
            Poll::Ready(Ok(()))
        } else {
            Poll::Pending
        }
    }
}

fn wait_id(id: i32) -> Result<(), Error> {
    let mut timeout = Some(time::Duration::new(0, 0));
    loop {
        let ids = events(timeout)?;
        process_event_ids(ids);
        if has_id(id) {
            break;
        }
        // the next call to events should block
        timeout = None;
    }
    Ok(())
}

#[cfg(all(target_arch = "wasm32"))]
impl io::Read for Conn {
    fn read(&mut self, buf: &mut [u8]) -> io::Result<usize> {
        remove_id(self.id);
        read(self.id, buf)
    }
}

#[cfg(all(target_arch = "wasm32"))]
impl io::Write for Conn {
    fn write(&mut self, buf: &[u8]) -> io::Result<usize> {
        write(self.id, buf)
    }
    fn flush(&mut self) -> io::Result<()> {
        Ok(())
    }
}

#[cfg(not(target_arch = "wasm32"))]
impl io::Read for Conn {
    fn read(&mut self, buf: &mut [u8]) -> io::Result<usize> {
        read(self.id, buf)
    }
}

#[cfg(not(target_arch = "wasm32"))]
impl io::Write for Conn {
    fn write(&mut self, buf: &[u8]) -> io::Result<usize> {
        write(self.id, buf)
    }
    fn flush(&mut self) -> io::Result<()> {
        Ok(())
    }
}

#[cfg(all(target_arch = "wasm32"))]
#[link(wasm_import_module = "embly")]
extern "C" {
    fn _read(id: i32, payload: *const u8, payload_len: u32, ln: *mut i32) -> u16;
    fn _write(id: i32, payload: *const u8, payload_len: u32, ln: *mut i32) -> u16;
    fn _spawn(name: *const u8, name_len: u32, id: *mut i32) -> u16;
    fn _events(
        non_blocking: u8,
        timeout_s: u64,
        timeout_ns: u32,
        ids: *const i32,
        ids_len: u32,
        ln: *mut i32,
    ) -> u16;
}

#[cfg(not(target_arch = "wasm32"))]
unsafe fn _events(
    _non_blocking: u8,
    _timeout_s: u64,
    _timeout_ns: u32,
    _ids: *const i32,
    _ids_len: u32,
    _ln: *mut i32,
) -> u16 {
    0
}

#[cfg(not(target_arch = "wasm32"))]
unsafe fn _read(_id: i32, _payload: *const u8, _payload_len: u32, ln: *mut i32) -> u16 {
    // lie and say no bytes
    *ln = 0;
    0
}

#[cfg(not(target_arch = "wasm32"))]
unsafe fn _write(_id: i32, _payload: *const u8, payload_len: u32, ln: *mut i32) -> u16 {
    // lie and say we write things
    *ln = payload_len as i32;
    0
}

#[cfg(not(target_arch = "wasm32"))]
unsafe fn _spawn(_name: *const u8, _name_len: u32, id: *mut i32) -> u16 {
    // lie and say the spawn id is 1
    *id = 1;
    0
}

fn read(id: i32, payload: &mut [u8]) -> io::Result<usize> {
    let mut ln: i32 = 0;
    let ln_ptr: *mut i32 = &mut ln;
    error::wasi_err_to_io_err(unsafe {
        _read(id, payload.as_ptr(), payload.len() as u32, ln_ptr)
    })?;
    Ok(ln as usize)
}

fn write(id: i32, payload: &[u8]) -> io::Result<usize> {
    let mut ln: i32 = 0;
    let ln_ptr: *mut i32 = &mut ln;
    error::wasi_err_to_io_err(unsafe {
        _write(id, payload.as_ptr(), payload.len() as u32, ln_ptr)
    })?;
    Ok(ln as usize)
}

fn spawn(name: &str) -> Result<i32, Error> {
    let mut id: i32 = 0;
    let id_ptr: *mut i32 = &mut id;
    error::wasi_err_to_io_err(unsafe { _spawn(name.as_ptr(), name.len() as u32, id_ptr) })?;
    Ok(id)
}

fn events(timeout: Option<time::Duration>) -> Result<Vec<i32>, Error> {
    let mut ln: i32 = 0;
    let ln_ptr: *mut i32 = &mut ln;
    let out: [i32; 10] = [0; 10];
    let mut timeout_s: u64 = 0;
    let mut timeout_ns: u32 = 0;
    let mut non_blocking: u8 = 0;
    if let Some(dur) = timeout {
        timeout_s = dur.as_secs();
        timeout_ns = dur.subsec_nanos();
    } else {
        non_blocking = 1
    };
    error::wasi_err_to_io_err(unsafe {
        _events(
            non_blocking,
            timeout_s,
            timeout_ns,
            out.as_ptr(),
            out.len() as u32,
            ln_ptr,
        )
    })?;
    Ok(out[..(ln as usize)].to_vec())
}

/// Run a Function
///
/// ```
/// use embly::{
///     Error,
///     prelude::*,
/// };
///
/// fn execute(mut conn: embly::Conn) -> Result<(), Error> {
///     conn.write_all(b"Hello\n")?;
///     let mut out = Vec::new();
///     conn.read_to_end(&mut out)?;
///     println!("{:?}", out);
///     Ok(())
/// }
/// async fn run(conn: embly::Conn) {
///     match execute(conn) {
///        Ok(_) => {}
///        Err(err) => {println!("got error: {}", err)}
///     }
/// }
/// fn main() {
///     embly::run(run);
/// }
/// ```
pub fn run<F>(to_run: fn(Conn) -> F)
where
    F: Future<Output = ()> + 'static,
{
    let c = Conn::new(1);
    task::Task::spawn(Box::pin(to_run(c)));
}

/// Run a function and panic on error
/// ```
/// use embly::{
///     Error,
///     prelude::*,
/// };
///
/// async fn execute(mut conn: embly::Conn) -> Result<(), Error> {
///     conn.write_all(b"Hello\n")?;
///     let mut out = Vec::new();
///     conn.read_to_end(&mut out)?;
///     println!("{:?}", out);
///     Ok(())
/// }
/// fn main() {
///     embly::run_catch_error(execute);
/// }
/// ```
pub fn run_catch_error<F>(to_run: fn(Conn) -> F)
where
    F: Future<Output = Result<(), Error>> + 'static,
{
    let c = Conn::new(1);
    task::Task::spawn(Box::pin(async move {
        match to_run(c).await {
            Ok(_) => {}
            Err(err) => println!("got error: {}", err),
        };
    }));
}
