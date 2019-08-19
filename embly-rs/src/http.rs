//! Tools for running http embly functions
//!
//! Use the http module to run functions that expect to recieve an http
//! response body and return an http response.
//!
//! ```no_run
//! use embly::Result;
//! use embly::http::{run, Body, Request, ResponseWriter};
//! use std::io::Write;
//!
//! fn execute(_req: Request<Body>, w: &mut ResponseWriter) -> Result<()> {
//!     w.write("hello world\n".as_bytes())?;
//!     w.status("200")?;
//!     w.header("Content-Length", "12")?;
//!     w.header("Content-Type", "text/plain")?;
//!     Ok(())
//! }
//!
//! fn main() -> Result<()> {
//!     run(execute)
//! }
//! ```
//!

use crate::error::{Error, Result};
use crate::Conn;
use http;
use http::header::{HeaderName, HeaderValue};
use http::response::Parts;
use http::status::StatusCode;
use http::HttpTryFrom;
pub use http::Request;
pub use http::Response;
use httparse;
use std::io;
use std::io::Read;
use std::io::Write;

/// An http body
#[derive(Debug)]
pub struct Body {
    conn: Conn,
    read_buf: Vec<u8>,
}

/// An http response writer. Used to write an http response, can either be used to write
/// a complete response, or stream a response
pub struct ResponseWriter {
    body: Body,
    parts: Parts,
    write_buf: Vec<u8>,
    // TODO: consolidate this logic
    headers_written: bool,
    function_returned: bool,
}

impl ResponseWriter {
    fn new(body: Body) -> Self {
        let (p, _) = Response::new(()).into_parts();
        Self {
            body,
            headers_written: false,
            function_returned: false,
            parts: p,
            write_buf: Vec::new(),
        }
    }
    fn write_headers(&mut self) -> Vec<u8> {
        let mut dst: Vec<u8> = Vec::new();

        if self.function_returned && !self.parts.headers.contains_key("Transfer-Encoding") {
            self.header("Content-Length", self.write_buf.len()).unwrap();
        }

        let init_cap = 30 + self.parts.headers.len() * AVERAGE_HEADER_SIZE;
        dst.reserve(init_cap);
        extend(&mut dst, b"HTTP/1.1 "); // todo: support passed version
        extend(&mut dst, self.parts.status.as_str().as_bytes());
        extend(&mut dst, b" ");
        extend(
            &mut dst,
            self.parts
                .status
                .canonical_reason()
                .unwrap_or("<none>")
                .as_bytes(),
        );
        extend(&mut dst, b"\r\n");
        for (name, values) in self.parts.headers.drain() {
            for value in values {
                extend(&mut dst, name.as_str().as_bytes());
                extend(&mut dst, b": ");
                extend(&mut dst, value.as_bytes());
                extend(&mut dst, b"\r\n");
            }
        }
        extend(&mut dst, b"\r\n");
        dst
    }
    /// add a header to this response
    pub fn header<K, V>(&mut self, key: K, value: V) -> Result<()>
    where
        HeaderName: HttpTryFrom<K>,
        HeaderValue: HttpTryFrom<V>,
    {
        match HeaderName::try_from(key) {
            Ok(key) => match HeaderValue::try_from(value) {
                Ok(value) => {
                    self.parts.headers.insert(key, value);
                    Ok(())
                }
                Err(e) => Err(Error::Http(e.into())),
            },
            Err(e) => Err(Error::Http(e.into())),
        }
    }
    /// set the http status for the response
    pub fn status<T>(&mut self, status: T) -> Result<()>
    where
        StatusCode: HttpTryFrom<T>,
    {
        match StatusCode::try_from(status) {
            Ok(s) => {
                self.parts.status = s;
                Ok(())
            }
            Err(e) => Err(Error::Http(e.into())),
        }
    }
}

impl io::Write for ResponseWriter {
    // think about this...
    // https://github.com/golang/go/blob/20e4540e90/src/net/http/server.go#L367-L390
    fn write(&mut self, buf: &[u8]) -> io::Result<usize> {
        // todo: should check existing headers and see if we can start sending a response
        self.write_buf.write(buf)
    }
    fn flush(&mut self) -> io::Result<()> {
        Ok(())
    }
}

/// flusher
pub trait Flusher {
    /// flusher
    fn flush_response(&mut self) -> Result<()>;
}

impl Flusher for ResponseWriter {
    fn flush_response(&mut self) -> Result<()> {
        // quick extra allocation for now to ensure flush makes one write call with
        // all bytes
        // TODO: remove extra allocation
        let mut out = Vec::new();
        if !self.headers_written {
            let dst = &self.write_headers();
            self.headers_written = true;
            out.write_all(dst)?;
        }
        out.write_all(&self.write_buf)?;
        self.body.write_all(&out)?;
        self.write_buf.clear();
        Ok(())
    }
}

#[inline]
fn extend(dst: &mut Vec<u8>, data: &[u8]) {
    dst.extend_from_slice(data);
}

impl io::Read for Body {
    fn read(&mut self, buf: &mut [u8]) -> io::Result<usize> {
        if !self.read_buf.is_empty() {
            let ln = (&self.read_buf[..]).read(buf)?;
            self.read_buf.truncate(self.read_buf.len() - ln);
            Ok(ln)
        } else {
            self.conn.read(buf)
        }
    }
}

impl io::Write for Body {
    fn write(&mut self, buf: &[u8]) -> io::Result<usize> {
        // todo: we're not just writing back to connection here
        // when we write to this message we're writing the
        // body so we need to be sure
        self.conn.write(buf)
    }
    fn flush(&mut self) -> io::Result<()> {
        Ok(())
    }
}

fn build_request_from_comm(c: Conn) -> Result<Request<Body>> {
    c.wait()?;
    reader_to_request(c)
}

// https://github.com/hyperium/hyper/blob/da9b0319ef8d85662f66ac3bea74036b3dd3744e/src/proto/h1/role.rs#L18
const MAX_HEADERS: usize = 100;
const AVERAGE_HEADER_SIZE: usize = 30;

fn reader_to_request<R: Read>(mut c: R) -> Result<Request<Body>> {
    let mut headers: Vec<httparse::Header> = vec![httparse::EMPTY_HEADER; MAX_HEADERS];
    let mut buf: Vec<u8> = Vec::new();

    let mut req = httparse::Request::new(&mut headers);
    c.read_to_end(&mut buf)?;
    let result = req.parse(&buf)?;
    if result.is_partial() {
        return Err(Error::InvalidHttpRequest);
    }
    let mut request = Request::builder();
    if let Some(uri) = req.path {
        request.uri(uri);
    }
    if let Some(method) = req.method {
        request.method(method);
    }
    // todo: reserve correct header capacity
    for h in &headers {
        if h.name.is_empty() && h.value.is_empty() {
            break;
        }
        request.header(h.name, h.value);
    }
    Ok(request.body(Body {
        conn: Conn { id: 0, buf: None },
        read_buf: buf[result.unwrap()..].to_vec(),
    })?)
}

// Will be used, currently just used for tests
#[allow(dead_code)]
fn reader_to_response<R: Read>(mut c: R) -> Result<Response<Body>> {
    let mut headers: Vec<httparse::Header> = vec![httparse::EMPTY_HEADER; MAX_HEADERS];
    let mut buf: Vec<u8> = Vec::new();

    let mut res = httparse::Response::new(&mut headers);
    c.read_to_end(&mut buf)?;
    let result = res.parse(&buf)?;
    if result.is_partial() {
        return Err(Error::InvalidHttpRequest);
    }

    let mut response = Response::builder();
    if let Some(code) = res.code {
        response.status(code);
    }

    // TODO: reason? version?

    // todo: reserve correct header capacity
    for h in &headers {
        if h.name.is_empty() && h.value.is_empty() {
            break;
        }
        response.header(h.name, h.value);
    }
    Ok(response.body(Body {
        conn: Conn { id: 0, buf: None },
        read_buf: buf[result.unwrap()..].to_vec(),
    })?)
}

/// Run an http handler Function
///
/// ```no_run
/// use embly::Result;
/// use embly::http::{run,Body, Request, ResponseWriter};
/// use std::io::Write;
///
/// fn execute(_req: Request<Body>, w: &mut ResponseWriter) -> Result<()> {
///     w.write("hello world\n".as_bytes())?;
///     w.status("200")?;
///     w.header("Content-Length", "12")?;
///     w.header("Content-Type", "text/plain")?;
///     Ok(())
/// }
///
/// fn main() -> Result<()> {
///     run(execute)
/// }
/// ```
pub fn run(to_run: fn(Request<Body>, &mut ResponseWriter) -> Result<()>) -> Result<()> {
    let function_id = 1;
    let c = Conn {
        id: function_id,
        buf: None,
    };
    let r = build_request_from_comm(c)?;
    let mut resp = ResponseWriter::new(Body {
        conn: Conn {
            id: function_id,
            buf: None,
        },
        read_buf: Vec::new(),
    });
    to_run(r, &mut resp)?;
    resp.function_returned = true;
    resp.flush_response()?;
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn basic_parse() {
        let b = "hello";
        let e = reader_to_request(b.as_bytes()).err().unwrap();
        match e {
            Error::InvalidHttpRequest => {}
            _ => panic!("{:?} should have recieved an invalid http request error", e),
        }
    }

    #[test]
    fn simple_valid_request() -> Result<()> {
        let b = "GET /c HTTP/1.1\r\nHost: f\r\n\r\n".as_bytes();
        reader_to_request(b)?;
        Ok(())
    }

    #[test]
    fn post_request() -> Result<()> {
        let b = "POST /test HTTP/1.1
Host: foo.example
Content-Type: application/x-www-form-urlencoded
Content-Length: 27

field1=value1&field2=value2";
        let mut request = reader_to_request(b.as_bytes())?;
        let body = request.body_mut();
        let mut b: Vec<u8> = Vec::new();
        body.read_to_end(&mut b)?;
        let values = "field1=value1&field2=value2";
        assert_eq!(b, values.as_bytes());
        Ok(())
    }

    fn test_response_writer() -> ResponseWriter {
        ResponseWriter::new(Body {
            conn: Conn {
                id: 0,
                buf: Some(Vec::new()),
            },
            read_buf: Vec::new(),
        })
    }

    #[test]
    fn basic_response() -> Result<()> {
        let mut w = test_response_writer();

        w.write_all(b"hello\n")?;
        w.status(401)?;
        w.header("Content-Type", "text/plain")?;

        w.function_returned = true;
        w.flush_response()?;
        let res = reader_to_response(w.body.conn)?;
        assert_eq!("6", res.headers().get("Content-Length").unwrap());
        assert_eq!(401, res.status());
        assert_eq!("text/plain", res.headers().get("Content-Type").unwrap());

        Ok(())
    }

    #[test]
    fn no_status_response() -> Result<()> {
        let mut w = test_response_writer();

        w.header("Content-Type", "text/plain")?;
        w.write_all(b"hello\n")?;
        w.function_returned = true;
        w.flush_response()?;

        let res = reader_to_response(w.body.conn)?;
        assert_eq!(200, res.status());
        Ok(())
    }

}
