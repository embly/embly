extern crate embly;

use embly::http;
use embly::http::{Body, Request, ResponseWriter};
use embly::kv;
use embly::prelude::*;
use embly::Error;
use std::time;

async fn execute(_req: Request<Body>, mut w: ResponseWriter) {
    match execute_async(&mut w).await {
        Ok(_) => {}
        Err(err) => {
            w.status(500).ok();
            w.write(format!("{:?}", err).as_bytes()).ok();
        }
    }
}

async fn execute_async(w: &mut ResponseWriter) -> Result<(), Error> {
    for _ in 0..10 {
        kv::set("hi".as_bytes(), "ho".as_bytes()).await?;
        let result = kv::get("hi".as_bytes()).await?;
        w.write(&result)?;
    }

    w.header("Content-Type", "text/plain")?;
    Ok(())
}

fn main() {
    http::run(execute);
}
