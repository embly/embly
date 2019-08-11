extern crate embly;

use embly::error::Result;
use embly::http;
use embly::http::{Body, Request, ResponseWriter};
use std::io::Read;
use std::io::Write;

fn execute(_req: Request<Body>, w: &mut ResponseWriter) -> Result<()> {
    let mut hello = embly::Function::spawn("hello")?;
    let mut buffer = Vec::new();
    hello.wait();
    hello.read_to_end(&mut buffer)?;
    println!("buffer contents after wait {:?}", buffer);
    w.write(&buffer)?;
    w.status("200")?;
    w.header("Content-Length", "6")?;
    w.header("Content-Type", "text/plain")?;
    Ok(())
}
fn main() -> Result<()> {
    http::run(execute)
}
