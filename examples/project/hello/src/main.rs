extern crate embly;

use embly::prelude::*;
use embly::Error;

fn execute(mut conn: embly::Conn) -> Result<(), Error> {
    conn.write_all(b"Embly is it!?\n")?;
    Ok(())
}

fn main() -> Result<(), Error> {
    embly::run(execute)
}
