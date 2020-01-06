use crate::http_proto::httpproto::Http;
use crate::Error;
use quick_protobuf::reader::deserialize_from_slice;
use quick_protobuf::writer::serialize_into_vec;
use std::io;
use std::io::Read;
use std::io::Write;

fn as_u32_le(array: &[u8]) -> u32 {
    u32::from(array[0])
        | (u32::from(array[1]) << 8)
        | (u32::from(array[2]) << 16)
        | (u32::from(array[3]) << 24)
}
fn u32_as_u8_le(x: u32) -> [u8; 4] {
    [
        (x & 0xff) as u8,
        ((x >> 8) & 0xff) as u8,
        ((x >> 16) & 0xff) as u8,
        ((x >> 24) & 0xff) as u8,
    ]
}

fn u64_as_u8_le(x: u64) -> [u8; 8] {
    [
        (x & 0xff) as u8,
        ((x >> 8) & 0xff) as u8,
        ((x >> 16) & 0xff) as u8,
        ((x >> 24) & 0xff) as u8,
        ((x >> 32) & 0xff) as u8,
        ((x >> 40) & 0xff) as u8,
        ((x >> 48) & 0xff) as u8,
        ((x >> 56) & 0xff) as u8,
    ]
}

pub fn write_msg<W: io::Write>(stream: &mut W, msg: Http) -> Result<(), Error> {
    let msg_bytes = serialize_into_vec(&msg)?;
    stream.write_all(&u32_as_u8_le(msg_bytes.len() as u32))?;
    stream.write_all(&msg_bytes)?;
    Ok(())
}

pub fn next_message<R: io::Read>(stream: &mut R) -> Result<Http, Error> {
    let mut size_bytes: [u8; 4] = [0; 4];
    stream.read_exact(&mut size_bytes)?;
    let size = as_u32_le(&size_bytes) as usize;
    let mut read = 0;
    if size == 0 {
        return Ok(Http::default());
    }
    let mut msg_bytes = vec![0; size];
    loop {
        let ln = stream.read(&mut msg_bytes[read..])?;
        read += ln;
        if ln == 0 || read == size {
            break;
        }
    }
    let msg: Http = deserialize_from_slice(&msg_bytes)?;
    Ok(msg)
}
