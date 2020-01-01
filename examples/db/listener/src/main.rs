extern crate embly;

use embly::http;
use embly::http::{Body, Request, ResponseWriter};
use embly::prelude::*;
use embly::Error;
use protos::data::User;
use std::time;
use vinyl_embly::query::field;
use vinyl_embly::DB;
mod protos;

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
    let db = DB::new("main").await?;
    let mut user = User::new();
    user.set_id(40);
    user.set_email("max@max.com".to_string());

    let insertion_resp = db.insert(user);
    let query_future = db.execute_query::<User>(field("id").equals(40 as i64));

    let first_user_future = db.load_record::<User, i64>(40);
    let _second_user_future = db.load_record::<User, i64>(30);

    let users = query_future.await?;
    w.write_all(format!("{:?}", users).to_string().as_bytes())?;

    println!("{:?}", first_user_future.await?);

    insertion_resp.await?;
    // second_user_future.await?;

    w.header("Content-Type", "text/plain")?;
    Ok(())
}

fn main() {
    http::run_async(execute);
}
