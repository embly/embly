extern crate embly;

use std::io;
use std::io::Read;
use std::io::Write;

fn execute(mut comm: embly::Comm) -> io::Result<()> {
    println!("Hello world!");
    comm.write_all(b"Hello\n")?;
    let mut out = Vec::new();
    comm.read_to_end(&mut out)?;
    println!("{:?}", out);
    Ok(())
}
fn main() {
    embly::run(execute);
}
