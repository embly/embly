extern crate embly;

use std::io;
use std::io::Read;
use std::io::Write;

fn execute(mut comm: embly::Comm) -> io::Result<()> {
    loop {
        comm.wait();
        let mut buffer = Vec::new();
        comm.read_to_end(&mut buffer)?;
        comm.write_all(&[&b"from embly: "[..], &buffer[..]].concat())?;
    }
    Ok(())
}
fn main() {
    embly::run(execute);
}
