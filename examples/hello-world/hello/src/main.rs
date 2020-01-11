extern crate embly;
use embly::{
  http::{Body, Request, ResponseWriter, run_catch_error},
  prelude::*,
  Error,
};

async fn execute(_req: Request<Body>, w: &mut ResponseWriter) -> Result<(), Error>{
    w.write_all(b"Hello World")?; // writing our hello response bytes 
    Ok(()) // if an error is returned the server will respond with an HTTP error
}

// this function is run first
fn main() {
    run_catch_error(execute); // this is the embly::http::run function that is specific to http responses
}
