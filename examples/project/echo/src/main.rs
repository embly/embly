extern crate embly;
extern crate rand;

use embly::prelude::*;
use embly::Error;
use std::time::SystemTime;

fn execute(mut conn: embly::Conn) -> Result<(), Error> {
    loop {
        conn.wait()?;
        let mut buffer = Vec::new();
        conn.read_to_end(&mut buffer)?;
        let now = SystemTime::now();
        let y = rand::random::<f64>();
        println!("time {:?} {}", now, y);
        conn.write_all(&[&b"from embly: "[..], &buffer[..]].concat())?;
    }
}

fn main() -> Result<(), Error> {
    embly::run(execute)
}
