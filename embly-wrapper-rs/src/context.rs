use {
    crate::{
        bimap::BidirectionalMap,
        bytes::{as_u32_le, u32_as_u8_le},
        error::{Error, Result},
        protos::comms::Message,
    },
    log::debug,
    protobuf::{parse_from_bytes, Message as _},
    std::{
        cmp,
        collections::{HashMap, VecDeque},
        io::prelude::*,
        os::unix::net::UnixStream,
        sync::mpsc::Receiver,
        time,
    },
};

pub struct EmblyCtx {
    pub address_map: BidirectionalMap<i32, u64>,
    pub address_count: i32,
    pub address: u64,
    parent_address: u64,
    pending_events: Vec<i32>,
    receiver: Receiver<Message>,
    read_buffers: HashMap<i32, VecDeque<Message>>,
    stream_writer: UnixStream,
}

impl EmblyCtx {
    pub fn new(
        receiver: Receiver<Message>,
        stream_writer: UnixStream,
        address: u64,
        parent_address: u64,
    ) -> Self {
        let address_map = BidirectionalMap::new();
        let mut ctx = Self {
            receiver,
            stream_writer,
            address_map,
            address,
            parent_address,
            address_count: 0,
            read_buffers: HashMap::new(),
            pending_events: Vec::new(),
        };
        ctx.add_address(parent_address);
        ctx
    }

    pub fn write(&mut self, id: i32, buf: &[u8]) -> Result<usize> {
        let mut msg = Message::new();
        msg.set_to(
            *self
                .address_map
                .get_value(id)
                .ok_or(Error::DescriptorDoesntExist)?,
        );
        msg.set_from(self.address);
        msg.set_data(buf.to_vec());
        self.write_msg(msg)?;
        Ok(buf.len())
    }

    pub fn read(&mut self, id: i32, buf: &mut [u8]) -> Result<usize> {
        self.process_messages(Some(time::Duration::new(0, 0)))?;

        if let Some(queue) = self.read_buffers.get_mut(&id) {
            if queue.is_empty() {
                return Ok(0);
            }
            let msg = queue.get_mut(0).expect("there should be something here");
            if msg.error != 0 {
                // TODO: actually create the correct io error
                return Err(Error::Io(std::io::Error::from(
                    std::io::ErrorKind::AddrNotAvailable,
                )));
            }
            let msg_data_ln = msg.get_data().len();
            let to_drain = cmp::min(buf.len(), msg_data_ln);
            let part: Vec<u8> = msg.mut_data().drain(..to_drain).collect();
            buf[..to_drain].copy_from_slice(&part);
            if msg.get_data().is_empty() {
                queue.pop_front();
            }
            Ok(part.len())
        } else {
            println!("no buffers for id");
            Ok(0)
        }
    }

    fn save_msg(&mut self, msg: Message) -> Result<i32> {
        if msg.from == 0 {
            debug!("message has invalid from of 0 {:?}", msg)
            // TODO: err
        }
        if msg.to == 0 {
            debug!("message has invalid to of 0 {:?}", msg)
            // TODO: err
        }

        let addr = self.add_address(msg.from);
        debug!("save_msg_addr {:?}", (addr, msg.from));
        if self.read_buffers.get(&addr).is_none() {
            self.read_buffers.insert(addr, VecDeque::new());
        }
        let buf = self.read_buffers.get_mut(&addr).unwrap();
        buf.push_back(msg);
        Ok(addr)
    }

    fn process_messages(&mut self, timeout: Option<time::Duration>) -> Result<()> {
        let mut new: Vec<Message> = self.receiver.try_iter().collect();

        // if we have events we return
        if new.is_empty() {
            if let Some(dur) = timeout {
                if let Ok(msg) = self.receiver.recv_timeout(dur) {
                    new.push(msg)
                }
            } else {
                // if no timeout is given we block forever
                let msg = self.receiver.recv()?;
                new.push(msg);
            }
        }
        for msg in new.drain(..) {
            let i = self.save_msg(msg)?;
            self.pending_events.push(i);
        }
        Ok(())
    }

    pub fn events_limited(
        &mut self,
        timeout: Option<time::Duration>,
        limit: usize,
    ) -> Result<Vec<i32>> {
        self.process_messages(timeout)?;
        let to_drain = cmp::min(self.pending_events.len(), limit);
        Ok(self.pending_events.drain(..to_drain).collect())
    }

    #[allow(dead_code)]
    pub fn events(&mut self, timeout: Option<time::Duration>) -> Result<Vec<i32>> {
        self.process_messages(timeout)?;
        Ok(self.pending_events.drain(..).collect())
    }

    fn add_address(&mut self, addr: u64) -> i32 {
        if let Some(k) = self.address_map.get_key(addr) {
            return *k;
        }
        self.address_count += 1;
        self.address_map.insert(self.address_count, addr);
        self.address_count
    }

    pub fn spawn(&mut self, name: &str) -> Result<i32> {
        let spawn_addr = rand::random::<u64>();
        let addr = self.add_address(spawn_addr);

        let mut msg = Message::new();
        msg.set_spawn(name.to_string());
        msg.set_to(self.parent_address);
        msg.set_from(self.address);

        msg.set_spawn_address(spawn_addr);

        // TODO! for now we generate the address ourselves here. This is just the easiest
        // because the function immediately knows where to send bytes to and the master
        // will receive events in order and be able to sort it out. Alternatively this
        // function would need be issued addresses to allocate or wait for a response

        self.write_msg(msg)?;
        Ok(addr)
    }
    fn write_msg(&mut self, msg: Message) -> Result<()> {
        write_msg(&mut self.stream_writer, msg)
    }
}

pub fn write_msg(stream: &mut UnixStream, msg: Message) -> Result<()> {
    let msg_bytes = msg.write_to_bytes()?;
    stream.write_all(&u32_as_u8_le(msg_bytes.len() as u32))?;
    stream.write_all(&msg_bytes)?;
    Ok(())
}

pub fn next_message(stream: &mut UnixStream) -> Result<Message> {
    let mut size_bytes: [u8; 4] = [0; 4];
    stream.read_exact(&mut size_bytes)?;
    let size = as_u32_le(&size_bytes) as usize;
    let mut read = 0;
    if size == 0 {
        return Ok(Message::new());
    }
    let mut msg_bytes = vec![0; size];
    loop {
        let ln = stream.read(&mut msg_bytes[read..])?;
        read += ln;
        debug!(
            "reading msg {:?}",
            (ln, msg_bytes[read..].len(), read, size)
        );
        if ln == 0 || read == size {
            break;
        }
    }
    let msg: Message = parse_from_bytes(&msg_bytes)?;
    Ok(msg)
}
