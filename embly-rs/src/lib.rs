//! Embly is a lightweight application runtime. It runs small isolated functions.
//! Functions can do a handful of things:
//!
//! - Receive bytes
//! - Send bytes
//! - Spawn a new function
//!
//! This library is used to access embly functionality from within a program
//! it is intended to only be built with `wasm32-wasi` but compilation should
//! work with other targets

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

pub use error::Result;
use std::io;
use std::time;

pub mod error;
pub mod http;

use std::sync::Mutex;
#[macro_use]
extern crate lazy_static;

lazy_static! {
    static ref EVENT_REGISTRY: Mutex<Vec<i32>> = {
        let m = Vec::new();
        Mutex::new(m)
    };
}
/// Conn's are connections and handle communication between functions or to a gateway
///
/// ## Receive Bytes
///
/// When a function begins execution it can optionally read in any bytes that it might have
/// been sent. Maybe there are bytes ready on startup, maybe it'll receive them later.
///
///
/// ```rust
/// use embly::Conn;
/// use embly::error::Result;
/// use std::io;
/// use std::io::Read;
///
/// fn entrypoint(mut conn: Conn) -> Result<()> {
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
/// Bytes can be written back. A spark is always executed by something. This could be a
/// command line call, a load balancer or another spark. Writing to a connection will send
/// those bytes back to the spark runner.
///
/// ```rust
/// use embly::Conn;
/// use std::io;
/// use std::io::Write;
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
pub struct Conn {
    id: i32,
}

/// Spawn a Function
///
/// ```
/// use embly::{Conn, spawn_function};
/// use embly::error::Result;
/// use std::io;
/// use std::io::Read;
/// use std::io::Write;
///
/// fn entrypoint(conn: Conn) -> Result<()> {
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
pub fn spawn_function(name: &str) -> Result<Conn> {
    Ok(Conn { id: spawn(name)? })
}

impl Conn {
    /// Wait for bytes to be available to be read from this Conn.
    pub fn wait(&self) -> Result<()> {
        let ids = events(Some(time::Duration::new(0, 0)))?;
        if ids.is_empty() {
            events(None)?;
        }
        Ok(())
    }
}

impl io::Read for Conn {
    fn read(&mut self, buf: &mut [u8]) -> io::Result<usize> {
        read(self.id, buf)
    }
}
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
    fn _read(id: i32, payload: *const u8, payload_len: u32, ln: *mut i32) -> u32;
    fn _write(id: i32, payload: *const u8, payload_len: u32, ln: *mut i32) -> u32;
    fn _spawn(name: *const u8, name_len: u32, id: *mut i32) -> u32;
    fn _events(
        non_blocking: u8,
        timeout_s: u64,
        timeout_ns: u32,
        ids: *const i32,
        ids_len: u32,
        ln: *mut i32,
    ) -> u32;
}

#[cfg(not(target_arch = "wasm32"))]
unsafe fn _events(
    _non_blocking: u8,
    _timeout_s: u64,
    _timeout_ns: u32,
    _ids: *const i32,
    _ids_len: u32,
    _ln: *mut i32,
) -> u32 {
    0
}

#[cfg(not(target_arch = "wasm32"))]
unsafe fn _read(_id: i32, _payload: *const u8, _payload_len: u32, ln: *mut i32) -> u32 {
    // lie and say EOF
    *ln = 0;
    0
}

#[cfg(not(target_arch = "wasm32"))]
unsafe fn _write(_id: i32, _payload: *const u8, payload_len: u32, ln: *mut i32) -> u32 {
    // lie and say we write things
    *ln = payload_len as i32;
    0
}

#[cfg(not(target_arch = "wasm32"))]
unsafe fn _spawn(_name: *const u8, _name_len: u32, id: *mut i32) -> u32 {
    *id = 1;
    0
}

fn read(id: i32, payload: &mut [u8]) -> io::Result<usize> {
    let mut ln: i32 = 0;
    let ln_ptr: *mut i32 = &mut ln;
    let _ = unsafe { _read(id, payload.as_ptr(), payload.len() as u32, ln_ptr) };
    Ok(ln as usize)
}

fn write(id: i32, payload: &[u8]) -> io::Result<usize> {
    let mut ln: i32 = 0;
    let ln_ptr: *mut i32 = &mut ln;
    let _ = unsafe { _write(id, payload.as_ptr(), payload.len() as u32, ln_ptr) };
    Ok(ln as usize)
}

fn spawn(name: &str) -> Result<i32> {
    let mut id: i32 = 0;
    let id_ptr: *mut i32 = &mut id;
    let _ = unsafe { _spawn(name.as_ptr(), name.len() as u32, id_ptr) };
    Ok(id)
}

fn events(timeout: Option<time::Duration>) -> Result<Vec<i32>> {
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
    let _ = unsafe {
        _events(
            non_blocking,
            timeout_s,
            timeout_ns,
            out.as_ptr(),
            out.len() as u32,
            ln_ptr,
        )
    };
    println!("events print  {:?}", (out, ln));
    Ok(out[..(ln as usize)].to_vec())
}

/// Run a Function
///
/// ```
///
/// use embly::Result;
/// use std::io;
/// use std::io::Read;
/// use std::io::Write;
///
/// fn execute(mut conn: embly::Conn) -> Result<()> {
///     conn.write_all(b"Hello\n")?;
///     let mut out = Vec::new();
///     conn.read_to_end(&mut out)?;
///     println!("{:?}", out);
///     Ok(())
/// }
/// fn main() -> Result<()> {
///     embly::run(execute)
/// }
/// ```
///
pub fn run(to_run: fn(Conn) -> Result<()>) -> Result<()> {
    println!("running regular func");
    let c = Conn { id: 1 };
    // todo: do something with this error
    to_run(c)
}

#[cfg(test)]
mod tests {
    #[test]
    fn it_works() {
        assert_eq!(2 + 2, 4);
    }
}
