extern crate embly;

use embly::prelude::*;

async fn execute(mut conn: embly::Conn) {
    conn.write_all(b"Embly is it!?\n").unwrap();
}

fn main() {
    embly::run(execute)
}
