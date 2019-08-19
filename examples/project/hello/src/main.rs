extern crate embly;

use embly::Result;
use std::io::Write;

fn execute(mut conn: embly::Conn) -> Result<()> {
    conn.write_all(b"Embly\n")?;
    Ok(())
}
fn main() -> Result<()> {
    embly::run(execute)
}
