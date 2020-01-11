extern crate embly;
use embly::{
  http::{run_catch_error, Body, Request, ResponseWriter},
  prelude::*,
  Error,
};

async fn execute(_req: Request<Body>, mut w: ResponseWriter) -> Result<(), Error> {
  w.write_all(b"Hello World")?; // writing our hello response bytes
  Ok(()) // if an error is returned the server will respond with an HTTP error
}

// this function is run first
fn main() {
  run_catch_error(execute); // this is the embly::http::run function that is specific to http responses
}
