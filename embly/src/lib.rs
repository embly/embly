//! embly is

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

use std::io;
use std::io::Result;

/// Comm is something without documentation yet
pub struct Comm {
    id: i32,
}

impl io::Read for Comm {
    fn read(&mut self, buf: &mut [u8]) -> Result<usize> {
        read(self.id, buf)
    }
}
impl io::Write for Comm {
    fn write(&mut self, buf: &[u8]) -> Result<usize> {
        write(self.id, buf)
    }
    fn flush(&mut self) -> Result<()> {
        Ok(())
    }
}

/// spawn a function!
pub struct Function {}

impl Function {}

extern "C" {
    fn _read(id: i32, payload: *const u8, payload_len: u32, ln: *const i32) -> u32;
    fn _write(id: i32, payload: *const u8, payload_len: u32, ln: *const i32) -> u32;
}

/// read something
pub fn read(id: i32, payload: &mut [u8]) -> Result<usize> {
    let ln: i32 = 0;
    let ln_ptr: *const i32 = &ln;
    let err = unsafe { _read(id, payload.as_ptr(), payload.len() as u32, ln_ptr) };
    println!("read err {:?}", err);
    Ok(ln as usize)
}

/// write something
pub fn write(id: i32, payload: &[u8]) -> Result<usize> {
    let ln: i32 = 0;
    let ln_ptr: *const i32 = &ln;
    let err = unsafe { _write(id, payload.as_ptr(), payload.len() as u32, ln_ptr) };
    println!("read err {:?}", err);
    Ok(ln as usize)
}
/// sfasDFaf
pub fn run(to_run: fn(Comm) -> io::Result<()>) {
    let c = Comm { id: 1 };
    to_run(c).unwrap();
    println!("i'm running are you")
}

#[no_mangle]
/// start
pub extern "C" fn start(_len: i32) {}

#[cfg(test)]
mod tests {
    #[test]
    fn it_works() {
        assert_eq!(2 + 2, 4);
    }
}
