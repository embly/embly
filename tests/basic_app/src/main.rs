
extern crate embly;

use embly::prelude::*;
use embly::Error;

async fn execute_async(mut conn: embly::Conn) {
    conn.write_all(b"ok then")
        .expect("should be able to write");
}

fn main() -> Result<(), Error> {
    embly::run(execute_async);
    Ok(())
}
