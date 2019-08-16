extern crate embly;

use std::io;
use std::io::Read;
use std::io::Write;

fn execute(mut conn: embly::Conn) -> io::Result<()> {
    loop {
        conn.wait();
        let mut buffer = Vec::new();
        conn.read_to_end(&mut buffer)?;
        conn.write_all(&[&b"from embly: "[..], &buffer[..]].concat())?;
    }
}
fn main() {
    embly::run(execute);
}
