// Automatically generated rust module for 'http.proto' file

#![allow(non_snake_case)]
#![allow(non_upper_case_globals)]
#![allow(non_camel_case_types)]
#![allow(unused_imports)]
#![allow(unknown_lints)]
#![allow(clippy)]
#![cfg_attr(rustfmt, rustfmt_skip)]


use std::io::Write;
use std::collections::HashMap;
use quick_protobuf::{MessageRead, MessageWrite, BytesReader, Writer, Result};
use quick_protobuf::sizeofs::*;
use super::*;

#[derive(Debug, Default, PartialEq, Clone)]
pub struct Http {
    pub proto_major: i32,
    pub proto_minor: i32,
    pub uri: String,
    pub headers: HashMap<String, String>,
    pub method: httpproto::mod_Http::Method,
    pub body: Vec<u8>,
}

impl<'a> MessageRead<'a> for Http {
    fn from_reader(r: &mut BytesReader, bytes: &'a [u8]) -> Result<Self> {
        let mut msg = Self::default();
        while !r.is_eof() {
            match r.next_tag(bytes) {
                Ok(8) => msg.proto_major = r.read_int32(bytes)?,
                Ok(16) => msg.proto_minor = r.read_int32(bytes)?,
                Ok(26) => msg.uri = r.read_string(bytes)?.to_owned(),
                Ok(34) => {
                    let (key, value) = r.read_map(bytes, |r, bytes| Ok(r.read_string(bytes)?.to_owned()), |r, bytes| Ok(r.read_string(bytes)?.to_owned()))?;
                    msg.headers.insert(key, value);
                }
                Ok(40) => msg.method = r.read_enum(bytes)?,
                Ok(50) => msg.body = r.read_bytes(bytes)?.to_owned(),
                Ok(t) => { r.read_unknown(bytes, t)?; }
                Err(e) => return Err(e),
            }
        }
        Ok(msg)
    }
}

impl MessageWrite for Http {
    fn get_size(&self) -> usize {
        0
        + if self.proto_major == 0i32 { 0 } else { 1 + sizeof_varint(*(&self.proto_major) as u64) }
        + if self.proto_minor == 0i32 { 0 } else { 1 + sizeof_varint(*(&self.proto_minor) as u64) }
        + if self.uri == String::default() { 0 } else { 1 + sizeof_len((&self.uri).len()) }
        + self.headers.iter().map(|(k, v)| 1 + sizeof_len(2 + sizeof_len((k).len()) + sizeof_len((v).len()))).sum::<usize>()
        + if self.method == httpproto::mod_Http::Method::GET { 0 } else { 1 + sizeof_varint(*(&self.method) as u64) }
        + if self.body == vec![] { 0 } else { 1 + sizeof_len((&self.body).len()) }
    }

    fn write_message<W: Write>(&self, w: &mut Writer<W>) -> Result<()> {
        if self.proto_major != 0i32 { w.write_with_tag(8, |w| w.write_int32(*&self.proto_major))?; }
        if self.proto_minor != 0i32 { w.write_with_tag(16, |w| w.write_int32(*&self.proto_minor))?; }
        if self.uri != String::default() { w.write_with_tag(26, |w| w.write_string(&**&self.uri))?; }
        for (k, v) in self.headers.iter() { w.write_with_tag(34, |w| w.write_map(2 + sizeof_len((k).len()) + sizeof_len((v).len()), 10, |w| w.write_string(&**k), 18, |w| w.write_string(&**v)))?; }
        if self.method != httpproto::mod_Http::Method::GET { w.write_with_tag(40, |w| w.write_enum(*&self.method as i32))?; }
        if self.body != vec![] { w.write_with_tag(50, |w| w.write_bytes(&**&self.body))?; }
        Ok(())
    }
}

pub mod mod_Http {


#[derive(Debug, PartialEq, Eq, Clone, Copy)]
pub enum Method {
    GET = 0,
    PUT = 1,
    POST = 2,
    DELETE = 3,
    PATCH = 4,
    OPTIONS = 5,
    TRACE = 6,
    CONNECT = 7,
}

impl Default for Method {
    fn default() -> Self {
        Method::GET
    }
}

impl From<i32> for Method {
    fn from(i: i32) -> Self {
        match i {
            0 => Method::GET,
            1 => Method::PUT,
            2 => Method::POST,
            3 => Method::DELETE,
            4 => Method::PATCH,
            5 => Method::OPTIONS,
            6 => Method::TRACE,
            7 => Method::CONNECT,
            _ => Self::default(),
        }
    }
}

impl<'a> From<&'a str> for Method {
    fn from(s: &'a str) -> Self {
        match s {
            "GET" => Method::GET,
            "PUT" => Method::PUT,
            "POST" => Method::POST,
            "DELETE" => Method::DELETE,
            "PATCH" => Method::PATCH,
            "OPTIONS" => Method::OPTIONS,
            "TRACE" => Method::TRACE,
            "CONNECT" => Method::CONNECT,
            _ => Self::default(),
        }
    }
}

}

