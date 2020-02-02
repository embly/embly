extern crate embly;
use embly::http::{Body, Request, ResponseWriter, run};
use embly::prelude::*;
use embly::Error;

fn execute(_req: Request<Body>, w: &mut ResponseWriter) -> Result<(), Error>{
    w.write_all(b"Hello World")?;
    Ok(())
}

fn main() -> Result<(), Error> {
    run(execute)
}