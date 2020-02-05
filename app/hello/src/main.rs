extern crate embly;
use embly::{
    http::{run_catch_error, Body, Request, ResponseWriter},
    prelude::*,
    Error,
};
use http::Method;

async fn execute(req: Request<Body>, mut w: ResponseWriter) -> Result<(), Error> {
    w.header("Access-Control-Allow-Origin", "*")?;
    if req.method() == Method::OPTIONS {
        w.status(200)?;
        w.header("Access-Control-Allow-Methods", "GET, OPTIONS")?;
        w.header("Access-Control-Allow-Headers", "Content-Type")?;
        return Ok(());
    }
    w.write_all(b"Hello World")?; // writing our hello response bytes
    Ok(()) // if an error is returned the server will respond with an HTTP error
}

// this function is run first
fn main() {
    run_catch_error(execute); // this is the embly::http::run function that is specific to http responses
}
