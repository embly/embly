extern crate embly;

use embly::http;
use embly::http::{Body, Request, ResponseWriter};
use embly::prelude::*;
use embly::Error;

fn execute(_req: Request<Body>, w: &mut ResponseWriter) -> Result<(), Error> {
    let mut hello = embly::spawn_function("hello")?;
    let mut buffer = Vec::new();
    hello.wait()?;
    hello.read_to_end(&mut buffer)?;
    println!("buffer contents after wait {:?}", buffer);
    w.write_all(&buffer)?;
    w.status("200")?;
    w.header("Content-Type", "text/plain")?;
    Ok(())
}

fn main() -> Result<(), Error> {
    http::run(execute)
}
