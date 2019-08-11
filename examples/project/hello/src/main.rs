extern crate embly;

use std::io;
use std::io::Write;

fn execute(mut comm: embly::Comm) -> io::Result<()> {
    comm.write_all(b"Embly\n")?;
    Ok(())
}
fn main() {
    embly::run(execute);
}
