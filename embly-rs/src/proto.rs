use crate::http_proto::httpproto::Http;
use crate::Conn;
use crate::Error;
use quick_protobuf::{BytesReader, MessageRead, MessageWrite, Writer};
use std::io;
use std::io::Read;

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

pub fn serialize(msg: &Http) -> Result<Vec<u8>, Error> {
    let mut v = Vec::with_capacity(msg.get_size());
    let mut writer = Writer::new(&mut v);
    msg.write_message(&mut writer)?;
    Ok(v)
}

pub fn deserialize(bytes: &Vec<u8>) -> Result<Http, Error> {
    let mut reader = BytesReader::from_bytes(&bytes);
    let out = Http::from_reader(&mut reader, &bytes)?;
    Ok(out)
}

pub fn write_msg<W: io::Write>(stream: &mut W, msg: Http) -> Result<(), Error> {
    let msg_bytes = serialize(&msg)?;
    stream.write_all(&u32_as_u8_le(msg_bytes.len() as u32))?;
    stream.write_all(&msg_bytes)?;
    Ok(())
}

pub fn next_message(stream: &mut Conn) -> Result<Http, Error> {
    let mut size_bytes: [u8; 4] = [0; 4];
    stream
        .read_exact(&mut size_bytes)
        .or_else(|err: io::Error| Err(err))?;
    let size = as_u32_le(&size_bytes) as usize;
    let mut read = 0;
    if size == 0 {
        return Ok(Http::default());
    }
    let mut msg_bytes = vec![0; size];
    loop {
        match stream.read(&mut msg_bytes[read..]) {
            Ok(ln) => {
                read += ln;
                if ln == 0 || read == size {
                    break;
                }
            }
            Err(_) => {
                stream.wait()?;
            }
        }
    }
    let msg: Http = deserialize(&msg_bytes)?;
    Ok(msg)
}
#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn test_proto() -> Result<(), Error> {
        let input: Vec<u8> = vec![
            8, 1, 16, 1, 26, 1, 47, 42, 24, 10, 4, 72, 111, 115, 116, 18, 16, 10, 14, 108, 111, 99,
            97, 108, 104, 111, 115, 116, 58, 56, 48, 56, 50, 42, 27, 10, 10, 85, 115, 101, 114, 45,
            65, 103, 101, 110, 116, 18, 13, 10, 11, 99, 117, 114, 108, 47, 55, 46, 54, 55, 46, 48,
            42, 15, 10, 6, 65, 99, 99, 101, 112, 116, 18, 5, 10, 3, 42, 47, 42,
        ];

        let msg: Http = deserialize(&input)?;
        println!("{:?}", msg);
        Ok(())
    }
}
