//! The embly wrapper runs webassembly code and managest embly runtime syscall functionality
//!

#![deny(
    trivial_numeric_casts,
    unstable_features,
    unused_extern_crates,
    unused_features
)]
#![warn(unused_import_braces, unused_parens)]
// #![deny(clippy::all)]

use error::Result;
use instance::Instance;
use lucet_runtime_internals::val;
use lucet_wasi;
use std::env;
use std::os::unix::net::UnixStream;

mod bimap;
mod bytes;
mod context;
mod error;
mod hostcalls;
mod instance;
mod protos;

fn main() -> Result<()> {
    env_logger::init();

    lucet_wasi::hostcalls::ensure_linked();
    lucet_runtime::lucet_internal_ensure_linked();

    let addr_string =
        env::var("EMBLY_ADDR").expect("EMBLY_ADDR environment variable should be available");
    let embly_module =
        env::var("EMBLY_MODULE").expect("EMBLY_MODULE environment variable should be available");
    let master_socket = UnixStream::connect("/tmp/embly.sock")?;
    let mut instance = Instance::new(embly_module, addr_string, master_socket)?;

    let exit_code = match instance
        .inst
        .run("main", &[val::Val::I32(0), val::Val::I32(0)])
    {
        // normal termination implies 0 exit code
        Ok(_) => 0,
        Err(lucet_runtime::Error::RuntimeTerminated(
            lucet_runtime::TerminationDetails::Provided(any),
        )) => *any
            .downcast_ref::<lucet_wasi::host::__wasi_exitcode_t>()
            .expect("termination yields an exitcode"),
        Err(e) => panic!("lucet-wasi runtime error: {}", e),
    };

    instance.send_exit_message(exit_code as i32)?;

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::context::next_message;
    use crate::context::EmblyCtx;
    use crate::protos::comms::Message;
    use log::debug;
    use std::os::unix::net::UnixStream;
    use std::str;
    use std::sync::mpsc;
    use std::sync::mpsc::channel;
    use std::time;

    const FUNC_ADDRESS: u64 = 8700;
    const MASTER: u64 = 8701;

    fn new_ctx() -> Result<(EmblyCtx, mpsc::Sender<Message>, UnixStream)> {
        let (sock1, sock2) = UnixStream::pair()?;
        let (sender, receiver) = channel();
        let ctx = EmblyCtx::new(receiver, sock1, FUNC_ADDRESS, MASTER);
        Ok((ctx, sender, sock2))
    }

    fn assert_send_and_read(
        id: i32,
        from: u64,
        to: u64,
        ctx: &mut EmblyCtx,
        sender: mpsc::Sender<Message>,
    ) -> Result<()> {
        let mut msg = Message::new();
        msg.set_data(b"hello".to_vec());
        msg.set_from(from);
        msg.set_to(to);
        sender.send(msg)?;

        let events = ctx.events(Some(time::Duration::new(0, 0)))?;
        debug!("{:?}", events);
        assert_eq!(1, events.len());
        let mut buf = vec![0; 4096];
        let ln = ctx.read(id, &mut buf)?;
        debug!("{}", ln);
        assert_eq!(str::from_utf8(&buf[..ln]).unwrap(), "hello");
        Ok(())
    }

    #[test]
    fn test_basic_read() -> Result<()> {
        let (mut ctx, sender, _stream) = new_ctx()?;

        assert_eq!(
            0,
            ctx.events(Some(time::Duration::new(0, 0))).unwrap().len()
        );

        assert_send_and_read(1, MASTER, FUNC_ADDRESS, &mut ctx, sender)?;
        Ok(())
    }
    #[test]
    fn test_spawn() -> Result<()> {
        let (mut ctx, sender, mut stream) = new_ctx()?;

        let addr = ctx.spawn("name")?;

        let msg = next_message(&mut stream)?;
        assert_eq!(msg.spawn, "name");
        let spawn_addr = msg.spawn_address;
        assert_eq!(msg.spawn_address, *ctx.address_map.get_value(addr).unwrap());

        assert_send_and_read(addr, spawn_addr, FUNC_ADDRESS, &mut ctx, sender)?;
        Ok(())
    }
}
