extern crate embly;

use embly::http;
use embly::http::{Body, Request, ResponseWriter};
use embly::prelude::*;
use embly::Error;
use protos::data::User;
use vinyl_embly::query::field;
use vinyl_embly::DB;

mod protos;

fn execute(_req: Request<Body>, w: &mut ResponseWriter) -> Result<(), Error> {
    // ::std::env::set_var("RUST_BACKTRACE", "full");

    let db = DB::new("main")?;
    let mut user = User::new();
    user.set_id(40);
    user.set_email("max@max.com".to_string());

    let mut insertion_resp = db.insert(user).expect("can insert");

    let mut users_response = db.execute_query::<User>(field("id").equals(40 as i64))?;

    // .execute_query(field("email").equals("max@max.com"))?

    let mut first_user_resp = db.load_record::<User, i64>(40)?;
    let _second_user_resp = db.load_record::<User, i64>(30)?;

    let users = users_response.wait()?;
    w.write_all(format!("{:?}", users).to_string().as_bytes())?;

    println!("{:?}", first_user_resp.wait()?);

    insertion_resp.wait()?;

    w.status("200")?;
    w.header("Content-Type", "text/plain")?;

    Ok(())
}

fn main() -> Result<(), Error> {
    http::run(execute)
}
