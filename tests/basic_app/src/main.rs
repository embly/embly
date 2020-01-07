
extern crate embly;

use embly::{prelude::*, spawn_function, Error};
use std::{time, thread};

async fn run_with_result(mut conn: embly::Conn) -> Result<(), Error> {
    let now = time::Instant::now();
    println!("starting {:?} s", time::Instant::now() - now);
    conn.write_all(b"ok then")?;

    let mut first = spawn_function("first")?;
    let mut second = spawn_function("second")?;

    first.await?;
    println!("first string {} {:?} s", first.string()?, time::Instant::now() - now);
    second.await?;
    println!("second string {} {:?} s", second.string()?, time::Instant::now() - now);
    let foo = 5;

    let mut bar = foo + 8;
    bar += 1;


    println!("{}", bar);

    Ok(())
}

async fn execute_async(conn: embly::Conn) {
    run_with_result(conn).await.expect("oh no");
}

fn main() -> Result<(), Error> {
    embly::run(execute_async);
    Ok(())
}    
