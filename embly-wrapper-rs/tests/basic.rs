extern crate embly_wrapper;
#[macro_use]
extern crate lazy_static;

use {
    embly_wrapper::{
        bytes::as_u32_le,
        context::{next_message, write_msg},
        error::Result,
        instance::Instance,
        protos::comms::Message,
    },
    lucet_runtime_internals::val,
    std::{
        fs::File,
        io::prelude::*,
        os::unix::net::UnixStream,
        path::{Path, PathBuf},
        process::Command,
        str,
        sync::Mutex,
        thread, time,
    },
};

lazy_static! {
    static ref BUILD_LOCK: Mutex<usize> = Mutex::new(0);
}

fn compile_and_create_instance(name: &str, code: &str) -> Result<(Instance, UnixStream)> {
    let _lock = BUILD_LOCK.lock();
    {
        let mut file = File::create("../tests/basic_app/src/main.rs")?;
        file.write_all(code.as_bytes())?;
        // file closes
    }

    // this makes this test work anywhere within the embly repo, but the rust wrapper
    // won't work without the cargo compiler flags within the wrapper project
    let basic_app_path = Path::join(
        std::env::current_exe()?.as_path(),
        PathBuf::from("../../../../tests/basic_app"),
    )
    .canonicalize()
    .unwrap();

    let output = Command::new("bash")
        .args(&[
            "-c",
            format!(
                "
    cd {} \
    && cargo +nightly build \
        --target wasm32-wasi \
        --release \
        -Z unstable-options \
        --out-dir ../scratch \
    && lucetc \
        --bindings ../../embly-wrapper-rs/bindings.json \
        --emit=so \
        --output ../scratch/{}.out \
        ../scratch/basic_app.wasm
        ",
                basic_app_path.as_path().display(),
                name
            )
            .as_str(),
        ])
        .output()?;
    if !output.status.success() {
        let output = str::from_utf8(&output.stderr).unwrap();
        println!("{}", output);
        panic!("build failed");
    }
    println!("Compilation complete for {}", name);
    let module = String::from(format!("../tests/scratch/{}.out", name));
    let addr = String::from("1001");
    let (mut sock1, sock2) = UnixStream::pair()?;

    let mut start_msg = Message::new();
    start_msg.set_parent_address(1000);
    start_msg.set_your_address(1001);
    write_msg(&mut sock1, start_msg)?;
    Ok((Instance::new(module, addr, sock2)?, sock1))
}

fn assert_addr(master_socket: &mut UnixStream) -> Result<()> {
    let mut size_bytes: [u8; 8] = [0; 8];
    master_socket.read_exact(&mut size_bytes)?;
    println!("{:?}", size_bytes);
    let addr = as_u32_le(&size_bytes) as usize;
    assert_eq!(addr, 1001);
    Ok(())
}

#[test]
fn test_async_basic() -> Result<()> {
    let val = "ok then";
    let code = r#"
extern crate embly;

use embly::prelude::*;
use embly::Error;

async fn execute_async(mut conn: embly::Conn) {
    conn.write_all(b"ok then")
        .expect("should be able to write");
}

fn main() -> Result<(), Error> {
    embly::run(execute_async);
    Ok(())
}
"#;

    let (mut instance, mut master_socket) = compile_and_create_instance("my_face", code)?;
    instance
        .inst
        .run("main", &[val::Val::I32(0), val::Val::I32(0)])?;
    instance.send_exit_message(0)?;

    assert_addr(&mut master_socket)?;

    let msg = next_message(&mut master_socket)?;
    assert_eq!(msg.data, val.as_bytes());

    Ok(())
}

fn send_with_delay(msg: Message, mut stream: UnixStream, delay_ms: u64) {
    thread::spawn(move || {
        thread::sleep(time::Duration::from_millis(delay_ms));
        write_msg(&mut stream, msg).expect("should send");
    });
}

#[test]
fn test_async_full() -> Result<()> {
    let code = r#"
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
"#;

    let (mut instance, mut master_socket) = compile_and_create_instance("async_full", code)?;
    thread::spawn(move || {
        assert_addr(&mut master_socket).expect("should");
        loop {
            if let Ok(msg) = next_message(&mut master_socket) {
                if msg.spawn_address == 0 {
                    continue;
                }
                println!("{:?}", msg);
                let mut resp_msg = Message::new();
                resp_msg.set_from(msg.spawn_address);
                resp_msg.set_to(msg.from);
                resp_msg.set_data(msg.spawn.as_bytes().to_vec());
                if msg.spawn == "first" {
                    send_with_delay(
                        resp_msg,
                        master_socket.try_clone().expect("should be able to clone"),
                        2000,
                    );
                } else if msg.spawn == "second" {
                    send_with_delay(
                        resp_msg,
                        master_socket.try_clone().expect("should be able to clone"),
                        1000,
                    );
                } else {
                    unreachable!("we only have two message types")
                }
            } else {
                println!("___ broken");
                break;
            }
        }
    });

    let now = time::Instant::now();
    instance
        .inst
        .run("main", &[val::Val::I32(0), val::Val::I32(0)])?;
    instance.send_exit_message(0)?;
    println!("run complete {:?}", time::Instant::now() - now);
    assert!((time::Instant::now() - now) < time::Duration::from_secs(3));
    Ok(())
}
