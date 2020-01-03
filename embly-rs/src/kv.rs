//! A simple key value store
//!
//! ```rust
//! use embly::kv;
//! use embly::Error;
//!
//! async fn entrypoint() -> Result<(), Error> {
//!     let key = b"key".to_vec();
//!     let value = b"value".to_vec();
//!     kv::set(&key, &value).await?;
//!     assert_eq!(value, kv::get(&key).await?);
//!     Ok(())
//! }

use crate::prelude::*;
use crate::{spawn_and_send, spawn_function};
use failure::{err_msg, Error};
use std::future::Future;

/// Set a binary key and value. Any existing value will be overwritten. Keys can
/// be no larger than 10,000kb and values can be no larger than 100,000kb
pub fn set(key: &[u8], value: &[u8]) -> impl Future<Output = Result<(), Error>> {
    let result = spawn_function("embly/kv/set");
    let write_result = write_key_and_value(key, value);
    async move {
        let mut conn = result?;
        let to_send = write_result?;
        conn.write(&to_send)?;
        Ok(())
    }
}
/// Get a key
pub fn get(key: &[u8]) -> impl Future<Output = Result<Vec<u8>, Error>> {
    let result = spawn_and_send("embly/kv/get", key);
    async move {
        let mut conn = result?;
        conn.await?;
        let mut out = Vec::new();
        conn.read_to_end(&mut out)?;
        Ok(out)
    }
}

fn u16_as_u8_le(x: u16) -> [u8; 2] {
    [(x & 0xff) as u8, ((x >> 8) & 0xff) as u8]
}

fn write_key_and_value(key: &[u8], value: &[u8]) -> Result<Vec<u8>, Error> {
    if key.len() > 10_000 {
        return Err(err_msg("key values can't be larger than 10,000"));
    }
    if value.len() > 100_000 {
        return Err(err_msg("value can't be greater than 100,000"));
    }
    let mut out = vec![0; key.len() + value.len() + 2];
    let key_len = key.len() as u16;
    out[..2].copy_from_slice(&u16_as_u8_le(key_len));
    out[2..key.len() + 2].copy_from_slice(key);
    out[key.len() + 2..].copy_from_slice(value);
    Ok(out)
}
