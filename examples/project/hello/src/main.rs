extern crate embly;

use std::io;
use std::io::Write;

fn execute(mut conn: embly::Conn) -> io::Result<()> {
    conn.write_all(b"Embly\n")?;
    Ok(())
}
fn main() {
    embly::run(execute);
}
