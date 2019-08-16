//! errors

use http;
use httparse;
use std::error;
use std::fmt;
use std::io;

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
