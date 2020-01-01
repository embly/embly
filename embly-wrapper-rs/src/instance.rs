use {
    crate::{
        bytes::u64_as_u8_le,
        context::{next_message, write_msg, EmblyCtx},
        error::{Error, Result},
        protos::comms::Message,
    },
    log::debug,
    lucet_runtime::{DlModule, Limits, MmapRegion, Module, Region},
    lucet_runtime_internals::{instance::InstanceHandle, module::ModuleInternal},
    lucet_wasi::WasiCtxBuilder,
    std::{
        io::prelude::*,
        os::unix::net::UnixStream,
        sync::{
            atomic::{AtomicBool, Ordering},
            mpsc::channel,
            Arc, Mutex,
        },
        thread,
    },
};

pub struct Instance {
    pub inst: InstanceHandle,
    parent_address: u64,
    your_address: u64,
    stream_closer: UnixStream,
    running: Arc<Mutex<AtomicBool>>,
}

impl Instance {
    pub fn new(
        embly_module: String,
        addr_string: String,
        mut master_socket: UnixStream,
    ) -> Result<Self> {
        let running = Arc::new(Mutex::new(AtomicBool::new(true)));
        let module = DlModule::load(&embly_module)?;
        // TODO: support memory constraints
        let min_globals_size = module.globals().len() * std::mem::size_of::<u64>();
        let globals_size = ((min_globals_size + 4096 - 1) / 4096) * 4096;
        let region = MmapRegion::create(
            1,
            &Limits {
                globals_size,
                heap_memory_size: 4_294_967_296,
                stack_size: 8_388_608,
                heap_address_space_size: 8_589_934_592,
            },
        )?;
        // TODO: likely remove inherit_env
        let ctx = WasiCtxBuilder::new()
            .inherit_stdio()
            .inherit_env()
            .env("RUST_BACKTRACE", "1");
        let socket_writer = master_socket.try_clone()?;
        let stream_closer = master_socket.try_clone()?;
        let (sender, receiver) = channel();
        let addr = addr_string.parse::<u64>().unwrap();
        master_socket.write_all(&u64_as_u8_le(addr))?;

        let thread_running = running.clone();
        thread::spawn(move || loop {
            debug!("reading bytes");
            if let Ok(msg) = next_message(&mut master_socket) {
                // channel has an infinite buffer, so this is where our messages go
                sender.send(msg).unwrap();
            } else {
                if thread_running.lock().unwrap().load(Ordering::Relaxed) {
                    panic!("error reading from socket, no longer listening");
                }
                break;
            }
        });
        let msg: Message = receiver.recv()?;
        debug!("got first message {:?}", msg);
        if msg.parent_address == 0 || msg.your_address == 0 {
            println!("either parent address or your address values are zero");
            return Err(Error::InvalidStartup(msg));
        }
        if msg.your_address != addr {
            panic!("addr doesn't match {} {}", addr, msg.your_address)
        }
        let parent_address = msg.parent_address;
        let your_address = msg.your_address;
        let embly_ctx = EmblyCtx::new(receiver, socket_writer, your_address, parent_address);
        let inst = region
            .new_instance_builder(module as Arc<dyn Module>)
            .with_embed_ctx(ctx.build().expect("WASI ctx can be created"))
            .with_embed_ctx(embly_ctx)
            .build()?;
        Ok(Instance {
            running,
            inst,
            parent_address,
            your_address,
            stream_closer,
        })
    }

    pub fn send_exit_message(&mut self, exit_code: i32) -> Result<()> {
        // inst.get_embed_ctx_mut()
        let mut msg = Message::new();
        msg.exit = exit_code; //todo: u32
        msg.exiting = true;
        msg.from = self.your_address;
        msg.to = self.parent_address;
        write_msg(&mut self.stream_closer, msg)?;
        self.running.lock().unwrap().store(false, Ordering::Relaxed);
        Ok(())
    }
}
