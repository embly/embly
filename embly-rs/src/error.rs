//! errors

use http;
use httparse;
use std::error;
use std::fmt;
use std::io;
use std::io::ErrorKind;

/// Result is the embly custom result type
pub type Result<T> = std::result::Result<T, Error>;

/// Error is an embly error
#[derive(Debug)]
pub enum Error {
    /// Bytes sent to this function were not a valid http request
    InvalidHttpRequest,
    /// Wrapper around http::Error
    Http(http::Error),
    /// Wrapper around io::Error
    Io(io::Error),
    /// Wrapper around httparse::Error
    HttpParse(httparse::Error),
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match *self {
            // todo: better error message
            Error::InvalidHttpRequest => write!(f, "The http request is invalid"),
            Error::Http(ref e) => e.fmt(f),
            Error::Io(ref e) => e.fmt(f),
            Error::HttpParse(ref e) => e.fmt(f),
        }
    }
}

impl error::Error for Error {
    fn source(&self) -> Option<&(dyn error::Error + 'static)> {
        match *self {
            Error::InvalidHttpRequest => None,
            Error::Http(ref e) => Some(e),
            Error::Io(ref e) => Some(e),
            Error::HttpParse(ref e) => Some(e),
        }
    }
}

impl From<io::Error> for Error {
    fn from(err: io::Error) -> Self {
        Error::Io(err)
    }
}

impl From<http::Error> for Error {
    fn from(err: http::Error) -> Self {
        Error::Http(err)
    }
}

impl From<httparse::Error> for Error {
    fn from(err: httparse::Error) -> Self {
        Error::HttpParse(err)
    }
}

/// convert a u16 wasi error type to an io::Error
pub fn wasi_err_to_io_err(err: u16) -> std::io::Result<()> {
    if err == 0 {
        Ok(())
    } else {
        Err(std::io::Error::from(match err {
            __WASI_ENOENT => ErrorKind::NotFound,
            __WASI_EACCES => ErrorKind::PermissionDenied,
            __WASI_ECONNREFUSED => ErrorKind::ConnectionRefused,
            __WASI_ECONNRESET => ErrorKind::ConnectionReset,
            __WASI_ECONNABORTED => ErrorKind::ConnectionAborted,
            __WASI_ENOTCONN => ErrorKind::NotConnected,
            __WASI_EADDRINUSE => ErrorKind::AddrInUse,
            __WASI_EADDRNOTAVAIL => ErrorKind::AddrNotAvailable,
            __WASI_EPIPE => ErrorKind::BrokenPipe,
            __WASI_EEXIST => ErrorKind::AlreadyExists,
            __WASI_EAGAIN => ErrorKind::WouldBlock,
            __WASI_EINVAL => ErrorKind::InvalidInput,
            __WASI_ETIMEDOUT => ErrorKind::TimedOut,
            __WASI_EINTR => ErrorKind::Interrupted,
            __WASI_EIO => ErrorKind::Other,
            _ => ErrorKind::Other,
        }))
    }
}

/// No error occurred. System call completed successfully.
pub const __WASI_ESUCCESS: u16 = 0;

/// Argument list too long.
pub const __WASI_E2BIG: u16 = 1;

/// Permission denied.
pub const __WASI_EACCES: u16 = 2;

/// Address in use.
pub const __WASI_EADDRINUSE: u16 = 3;

/// Address not available.
pub const __WASI_EADDRNOTAVAIL: u16 = 4;

/// Address family not supported.
pub const __WASI_EAFNOSUPPORT: u16 = 5;

/// Resource unavailable, or operation would block.
pub const __WASI_EAGAIN: u16 = 6;

/// Connection already in progress.
pub const __WASI_EALREADY: u16 = 7;

/// Bad file descriptor.
pub const __WASI_EBADF: u16 = 8;

/// Bad message.
pub const __WASI_EBADMSG: u16 = 9;

/// Device or resource busy.
pub const __WASI_EBUSY: u16 = 10;

/// Operation canceled.
pub const __WASI_ECANCELED: u16 = 11;

/// No child processes.
pub const __WASI_ECHILD: u16 = 12;

/// Connection aborted.
pub const __WASI_ECONNABORTED: u16 = 13;

/// Connection refused.
pub const __WASI_ECONNREFUSED: u16 = 14;

/// Connection reset.
pub const __WASI_ECONNRESET: u16 = 15;

/// Resource deadlock would occur.
pub const __WASI_EDEADLK: u16 = 16;

/// Destination address required.
pub const __WASI_EDESTADDRREQ: u16 = 17;

/// Mathematics argument out of domain of function.
pub const __WASI_EDOM: u16 = 18;

/// Reserved. (Quota exceeded.)
pub const __WASI_EDQUOT: u16 = 19;

/// File exists.
pub const __WASI_EEXIST: u16 = 20;

/// Bad address.
pub const __WASI_EFAULT: u16 = 21;

/// File too large.
pub const __WASI_EFBIG: u16 = 22;

/// Host is unreachable.
pub const __WASI_EHOSTUNREACH: u16 = 23;

/// Identifier removed.
pub const __WASI_EIDRM: u16 = 24;

/// Illegal byte sequence.
pub const __WASI_EILSEQ: u16 = 25;

/// Operation in progress.
pub const __WASI_EINPROGRESS: u16 = 26;

/// Interrupted function.
pub const __WASI_EINTR: u16 = 27;

/// Invalid argument.
pub const __WASI_EINVAL: u16 = 28;

/// I/O error.
pub const __WASI_EIO: u16 = 29;

/// Socket is connected.
pub const __WASI_EISCONN: u16 = 30;

/// Is a directory.
pub const __WASI_EISDIR: u16 = 31;

/// Too many levels of symbolic links.
pub const __WASI_ELOOP: u16 = 32;

/// File descriptor value too large.
pub const __WASI_EMFILE: u16 = 33;

/// Too many links.
pub const __WASI_EMLINK: u16 = 34;

/// Message too large.
pub const __WASI_EMSGSIZE: u16 = 35;

/// Reserved. (Multihop attempted.)
pub const __WASI_EMULTIHOP: u16 = 36;

/// Filename too long.
pub const __WASI_ENAMETOOLONG: u16 = 37;

/// Network is down.
pub const __WASI_ENETDOWN: u16 = 38;

/// Connection aborted by network.
pub const __WASI_ENETRESET: u16 = 39;

/// Network unreachable.
pub const __WASI_ENETUNREACH: u16 = 40;

/// Too many files open in system.
pub const __WASI_ENFILE: u16 = 41;

/// No buffer space available.
pub const __WASI_ENOBUFS: u16 = 42;

/// No such device.
pub const __WASI_ENODEV: u16 = 43;

/// No such file or directory.
pub const __WASI_ENOENT: u16 = 44;

/// Executable file format error.
pub const __WASI_ENOEXEC: u16 = 45;

/// No locks available.
pub const __WASI_ENOLCK: u16 = 46;

/// Reserved. (Link has been severed.)
pub const __WASI_ENOLINK: u16 = 47;

/// Not enough space.
pub const __WASI_ENOMEM: u16 = 48;

/// No message of the desired type.
pub const __WASI_ENOMSG: u16 = 49;

/// Protocol not available.
pub const __WASI_ENOPROTOOPT: u16 = 50;

/// No space left on device.
pub const __WASI_ENOSPC: u16 = 51;

/// Function not supported. (Always unsupported.)
pub const __WASI_ENOSYS: u16 = 52;

/// The socket is not connected.
pub const __WASI_ENOTCONN: u16 = 53;

/// Not a directory or a symbolic link to a directory.
pub const __WASI_ENOTDIR: u16 = 54;

/// Directory not empty.
pub const __WASI_ENOTEMPTY: u16 = 55;

/// State not recoverable.
pub const __WASI_ENOTRECOVERABLE: u16 = 56;

/// Not a socket.
pub const __WASI_ENOTSOCK: u16 = 57;

/// Not supported, or operation not supported on socket. (Transient unsupported.)
pub const __WASI_ENOTSUP: u16 = 58;

/// Inappropriate I/O control operation.
pub const __WASI_ENOTTY: u16 = 59;

/// No such device or address.
pub const __WASI_ENXIO: u16 = 60;

/// Value too large to be stored in data type.
pub const __WASI_EOVERFLOW: u16 = 61;

/// Previous owner died.
pub const __WASI_EOWNERDEAD: u16 = 62;

/// Operation not permitted.
pub const __WASI_EPERM: u16 = 63;

/// Broken pipe.
pub const __WASI_EPIPE: u16 = 64;

/// Protocol error.
pub const __WASI_EPROTO: u16 = 65;

/// Protocol not supported.
pub const __WASI_EPROTONOSUPPORT: u16 = 66;

/// Protocol wrong type for socket.
pub const __WASI_EPROTOTYPE: u16 = 67;

/// Result too large.
pub const __WASI_ERANGE: u16 = 68;

/// Read-only file system.
pub const __WASI_EROFS: u16 = 69;

/// Invalid seek.
pub const __WASI_ESPIPE: u16 = 70;

/// No such process.
pub const __WASI_ESRCH: u16 = 71;

/// Reserved. (Stale file handle.)
pub const __WASI_ESTALE: u16 = 72;

/// Connection timed out.
pub const __WASI_ETIMEDOUT: u16 = 73;

/// Text file busy.
pub const __WASI_ETXTBSY: u16 = 74;

/// Cross-device link.
pub const __WASI_EXDEV: u16 = 75;

/// Extension: Capabilities insufficient.
pub const __WASI_ENOTCAPABLE: u16 = 76;
