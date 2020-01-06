extern crate embly;

use embly::http;
use embly::http::{Body, Request, ResponseWriter};
use embly::prelude::*;
use embly::Error;

async fn execute(_req: Request<Body>, w: &mut ResponseWriter) -> Result<(), Error> {
    let mut hello = embly::spawn_function("hello")?;
    let mut buffer = Vec::new();
    hello.wait()?;
    hello.read_to_end(&mut buffer)?;
    w.write_all(&buffer)?;
    w.status("200")?;
    w.header("Content-Type", "text/plain")?;
    Ok(())
}

async fn catch_error(req: Request<Body>, mut w: ResponseWriter) {
    match execute(req, &mut w).await {
        Ok(_) => {}
        Err(err) => {
            w.status("500").unwrap();
            w.write(format!("{}", err).as_bytes()).unwrap();
        }
    };
}

fn main() {
    http::run(catch_error);
}
