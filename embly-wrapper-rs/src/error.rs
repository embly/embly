use crate::protos::comms::Message;
use lucet_runtime_internals;
use lucet_wasi::host;
use protobuf;
use std::error;
use std::fmt;
use std::io;
use std::io::ErrorKind;
use std::sync;

pub type Result<T> = std::result::Result<T, Error>;

#[derive(Debug)]
pub enum Error {
    DescriptorDoesntExist,
    InvalidStartup(Message),
    Io(io::Error),
    Proto(protobuf::error::ProtobufError),
    MpscRecv(sync::mpsc::RecvError),
    MpscSend(sync::mpsc::SendError<Message>),
    Lucet(lucet_runtime_internals::error::Error),
    WasiErr(u16),
}

impl Error {
    pub fn to_wasi_err(&self) -> u32 {
        match self {
            Error::Io(ref e) => match e.kind() {
                ErrorKind::NotFound => host::__WASI_ENOENT,
                ErrorKind::PermissionDenied => host::__WASI_EACCES,
                ErrorKind::ConnectionRefused => host::__WASI_ECONNREFUSED,
                ErrorKind::ConnectionReset => host::__WASI_ECONNRESET,
                ErrorKind::ConnectionAborted => host::__WASI_ECONNABORTED,
                ErrorKind::NotConnected => host::__WASI_ENOTCONN,
                ErrorKind::AddrInUse => host::__WASI_EADDRINUSE,
                ErrorKind::AddrNotAvailable => host::__WASI_EADDRNOTAVAIL,
                ErrorKind::BrokenPipe => host::__WASI_EPIPE,
                ErrorKind::AlreadyExists => host::__WASI_EEXIST,
                ErrorKind::WouldBlock => host::__WASI_EAGAIN,
                ErrorKind::InvalidInput | ErrorKind::InvalidData => host::__WASI_EINVAL,
                ErrorKind::TimedOut => host::__WASI_ETIMEDOUT,
                ErrorKind::Interrupted => host::__WASI_EINTR,
                ErrorKind::WriteZero | ErrorKind::Other | ErrorKind::UnexpectedEof | _ => {
                    host::__WASI_EIO
                }
            },
            Error::WasiErr(ref e) => u32::from(*e),
            _ => 0,
        }
    }
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match *self {
            Self::DescriptorDoesntExist => write!(f, "Id doesn't exist"),
            Self::InvalidStartup(ref msg) => write!(f, "Invalid startup message {:?}", msg),
            Self::Io(ref e) => e.fmt(f),
            Self::Proto(ref e) => e.fmt(f),
            Self::MpscRecv(ref e) => e.fmt(f),
            Self::MpscSend(ref e) => e.fmt(f),
            Self::Lucet(ref e) => e.fmt(f),
            Self::WasiErr(ref e) => e.fmt(f),
        }
    }
}

impl error::Error for Error {
    fn source(&self) -> Option<&(dyn error::Error + 'static)> {
        match *self {
            Self::DescriptorDoesntExist => None,
            Self::InvalidStartup(_) => None,
            Self::Io(ref e) => Some(e),
            Self::Proto(ref e) => Some(e),
            Self::MpscRecv(ref e) => Some(e),
            Self::MpscSend(ref e) => Some(e),
            Self::WasiErr(ref _e) => None,
            // can't figure out how to cast failure::Error to error::Error
            // Some(e.compat()) doesn't work
            Self::Lucet(ref _e) => None,
        }
    }
}

impl From<io::Error> for Error {
    fn from(err: io::Error) -> Self {
        Self::Io(err)
    }
}

impl From<u16> for Error {
    fn from(err: u16) -> Self {
        Self::WasiErr(err)
    }
}

impl From<sync::mpsc::RecvError> for Error {
    fn from(err: sync::mpsc::RecvError) -> Self {
        Self::MpscRecv(err)
    }
}
impl From<sync::mpsc::SendError<Message>> for Error {
    fn from(err: sync::mpsc::SendError<Message>) -> Self {
        Self::MpscSend(err)
    }
}
impl From<protobuf::error::ProtobufError> for Error {
    fn from(err: protobuf::error::ProtobufError) -> Self {
        Self::Proto(err)
    }
}

impl From<lucet_runtime_internals::error::Error> for Error {
    fn from(err: lucet_runtime_internals::error::Error) -> Self {
        Self::Lucet(err)
    }
}
