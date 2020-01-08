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
    pub status: i32,
    pub headers: HashMap<String, httpproto::HeaderList>,
    pub method: httpproto::mod_Http::Method,
    pub body: Vec<u8>,
    pub eof: bool,
}

impl<'a> MessageRead<'a> for Http {
    fn from_reader(r: &mut BytesReader, bytes: &'a [u8]) -> Result<Self> {
        let mut msg = Self::default();
        while !r.is_eof() {
            match r.next_tag(bytes) {
                Ok(8) => msg.proto_major = r.read_int32(bytes)?,
                Ok(16) => msg.proto_minor = r.read_int32(bytes)?,
                Ok(26) => msg.uri = r.read_string(bytes)?.to_owned(),
                Ok(32) => msg.status = r.read_int32(bytes)?,
                Ok(42) => {
                    let (key, value) = r.read_map(bytes, |r, bytes| Ok(r.read_string(bytes)?.to_owned()), |r, bytes| Ok(r.read_message::<httpproto::HeaderList>(bytes)?))?;
                    msg.headers.insert(key, value);
                }
                Ok(48) => msg.method = r.read_enum(bytes)?,
                Ok(58) => msg.body = r.read_bytes(bytes)?.to_owned(),
                Ok(64) => msg.eof = r.read_bool(bytes)?,
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
        + if self.status == 0i32 { 0 } else { 1 + sizeof_varint(*(&self.status) as u64) }
        + self.headers.iter().map(|(k, v)| 1 + sizeof_len(2 + sizeof_len((k).len()) + sizeof_len((v).get_size()))).sum::<usize>()
        + if self.method == httpproto::mod_Http::Method::GET { 0 } else { 1 + sizeof_varint(*(&self.method) as u64) }
        + if self.body == vec![] { 0 } else { 1 + sizeof_len((&self.body).len()) }
        + if self.eof == false { 0 } else { 1 + sizeof_varint(*(&self.eof) as u64) }
    }

    fn write_message<W: Write>(&self, w: &mut Writer<W>) -> Result<()> {
        if self.proto_major != 0i32 { w.write_with_tag(8, |w| w.write_int32(*&self.proto_major))?; }
        if self.proto_minor != 0i32 { w.write_with_tag(16, |w| w.write_int32(*&self.proto_minor))?; }
        if self.uri != String::default() { w.write_with_tag(26, |w| w.write_string(&**&self.uri))?; }
        if self.status != 0i32 { w.write_with_tag(32, |w| w.write_int32(*&self.status))?; }
        for (k, v) in self.headers.iter() { w.write_with_tag(42, |w| w.write_map(2 + sizeof_len((k).len()) + sizeof_len((v).get_size()), 10, |w| w.write_string(&**k), 18, |w| w.write_message(v)))?; }
        if self.method != httpproto::mod_Http::Method::GET { w.write_with_tag(48, |w| w.write_enum(*&self.method as i32))?; }
        if self.body != vec![] { w.write_with_tag(58, |w| w.write_bytes(&**&self.body))?; }
        if self.eof != false { w.write_with_tag(64, |w| w.write_bool(*&self.eof))?; }
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

#[derive(Debug, Default, PartialEq, Clone)]
pub struct HeaderList {
    pub header: Vec<String>,
}

impl<'a> MessageRead<'a> for HeaderList {
    fn from_reader(r: &mut BytesReader, bytes: &'a [u8]) -> Result<Self> {
        let mut msg = Self::default();
        while !r.is_eof() {
            match r.next_tag(bytes) {
                Ok(10) => msg.header.push(r.read_string(bytes)?.to_owned()),
                Ok(t) => { r.read_unknown(bytes, t)?; }
                Err(e) => return Err(e),
            }
        }
        Ok(msg)
    }
}

impl MessageWrite for HeaderList {
    fn get_size(&self) -> usize {
        0
        + self.header.iter().map(|s| 1 + sizeof_len((s).len())).sum::<usize>()
    }

    fn write_message<W: Write>(&self, w: &mut Writer<W>) -> Result<()> {
        for s in &self.header { w.write_with_tag(10, |w| w.write_string(&**s))?; }
        Ok(())
    }
}

