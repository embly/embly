//! Tools for running http embly functions
//!
//! Use the http module to run functions that expect to recieve an http
//! response body and return an http response.
//!
//! ```no_run
//! use embly::{
//!     Error,
//!     prelude::*,
//!     http::{Body, Request, ResponseWriter},
//! };
//!
//! async fn execute (_req: Request<Body>, w: &mut ResponseWriter) -> Result<(), Error> {
//!     w.write("hello world\n".as_bytes())?;
//!     w.status("200")?;
//!     w.header("Content-Length", "12")?;
//!     w.header("Content-Type", "text/plain")?;
//!     Ok(())
//! }
//! async fn catch_error(req: Request<Body>, mut w: ResponseWriter) {
//!     match execute(req, &mut w).await {
//!         Ok(_) => {}
//!         Err(err) => {
//!             w.status("500").unwrap();
//!             w.write(format!("{}", err).as_bytes()).unwrap();
//!         },
//!     };
//! }
//!
//! fn main() {
//!     ::embly::http::run(catch_error);
//! }
//! ```
//!

use crate::error::Error as EmblyError;
use crate::http_proto::httpproto::Http;
use crate::proto;
use crate::task;
use crate::Conn;
use failure::Error;
use http;
use http::header::{HeaderName, HeaderValue};
use http::response::Parts;
use http::status::StatusCode;
use http::HttpTryFrom;
pub use http::Request;
pub use http::Response;
use httparse;
use std::future::Future;
use std::io;
use std::io::Read;
use std::io::Write;
use std::sync::Arc;
use std::sync::Mutex;

/// An http body
#[derive(Debug)]
pub struct Body {
    conn: Conn,
    content_length: Option<usize>,
    read_buf: Vec<u8>,
}

struct Interior {
    body: Body,
    parts: Parts,
    write_buf: Vec<u8>,
}

impl Interior {
    fn header<K, V>(&mut self, key: K, value: V) -> Result<(), Error>
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
                Err(e) => Err(EmblyError::Http(e.into()).into()),
            },
            Err(e) => Err(EmblyError::Http(e.into()).into()),
        }
    }
}

/// An http response writer. Used to write an http response, can either be used to write
/// a complete response, or stream a response
///
/// ```rust
/// use embly::{
///     Error,
///     http::ResponseWriter,
///     prelude::*,
/// };
/// fn write_response(mut w: ResponseWriter) -> Result<(), Error> {
///     w.write("hello world\n".as_bytes())?;
///     w.status("200")?;
///     w.header("Content-Type", "text/plain")
/// }
/// ```
pub struct ResponseWriter {
    interior: Arc<Mutex<Interior>>,

    // TODO: consolidate this logic
    headers_written: bool,
    function_returned: bool,
}

impl Clone for ResponseWriter {
    fn clone(&self) -> Self {
        Self {
            headers_written: self.headers_written,
            function_returned: self.function_returned,
            interior: self.interior.clone(),
        }
    }
}

impl ResponseWriter {
    fn new(body: Body) -> Self {
        let (p, _) = Response::new(()).into_parts();
        Self {
            headers_written: false,
            function_returned: false,
            interior: Arc::new(Mutex::new(Interior {
                body: body,
                parts: p,
                write_buf: Vec::new(),
            })),
        }
    }
    fn write_headers(&mut self) -> Vec<u8> {
        let mut dst: Vec<u8> = Vec::new();

        let mut interior = self.interior.lock().unwrap();

        if self.function_returned && !interior.parts.headers.contains_key("Transfer-Encoding") {
            let len = interior.write_buf.len();
            interior.header("Content-Length", len).unwrap();
        }

        let init_cap = 30 + interior.parts.headers.len() * AVERAGE_HEADER_SIZE;
        dst.reserve(init_cap);
        extend(&mut dst, b"HTTP/1.1 "); // todo: support passed version
        extend(&mut dst, interior.parts.status.as_str().as_bytes());
        extend(&mut dst, b" ");
        extend(
            &mut dst,
            interior
                .parts
                .status
                .canonical_reason()
                .unwrap_or("<none>")
                .as_bytes(),
        );
        extend(&mut dst, b"\r\n");
        for (name, values) in interior.parts.headers.drain() {
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
    pub fn header<K, V>(&mut self, key: K, value: V) -> Result<(), Error>
    where
        HeaderName: HttpTryFrom<K>,
        HeaderValue: HttpTryFrom<V>,
    {
        self.interior.lock().unwrap().header(key, value)
    }
    /// set the http status for the response
    pub fn status<T>(&mut self, status: T) -> Result<(), Error>
    where
        StatusCode: HttpTryFrom<T>,
    {
        match StatusCode::try_from(status) {
            Ok(s) => {
                self.interior.lock().unwrap().parts.status = s;
                Ok(())
            }
            Err(e) => Err(EmblyError::Http(e.into()).into()),
        }
    }
}

impl io::Write for ResponseWriter {
    // think about this...
    // https://github.com/golang/go/blob/20e4540e90/src/net/http/server.go#L367-L390
    fn write(&mut self, buf: &[u8]) -> io::Result<usize> {
        // todo: should check existing headers and see if we can start sending a response
        self.interior.lock().unwrap().write_buf.write(buf)
    }
    fn flush(&mut self) -> io::Result<()> {
        Ok(())
    }
}

/// The flusher trait can flush an http response
pub trait Flusher {
    /// Flushes the current buffered content out as an http response.
    /// Can be used to stream back an http response.
    fn flush_response(&mut self) -> Result<(), Error>;
}

impl Flusher for ResponseWriter {
    fn flush_response(&mut self) -> Result<(), Error> {
        // quick extra allocation for now to ensure flush makes one write call with
        // all bytes
        // TODO: remove extra allocation
        let mut out = Vec::new();
        if !self.headers_written {
            let dst = &self.write_headers();
            self.headers_written = true;
            out.write_all(dst)?;
        }
        let mut interior = self.interior.lock().unwrap();
        out.write_all(&interior.write_buf)?;
        interior.body.write_all(&out)?;
        interior.write_buf.clear();
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
            self.read_buf.drain(0..ln);
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

fn build_request_from_comm(c: &mut Conn) -> Result<Request<Body>, Error> {
    c.wait()?;
    let id = c.id;
    let http = proto::next_message(c)?;
    let mut request = http_proto_to_request(http);
    let body = request.body_mut();
    body.conn.id = id;
    Ok(request)
}

// https://github.com/hyperium/hyper/blob/da9b0319ef8d85662f66ac3bea74036b3dd3744e/src/proto/h1/role.rs#L18
const MAX_HEADERS: usize = 100;
const AVERAGE_HEADER_SIZE: usize = 30;

fn http_proto_to_request(http: Http) -> Request<Body> {
    let mut request = Request::builder();
    request.uri(http.uri);
    // hardcode a map?
    request.method(format!("{:?}", http.method).as_str());
    for (h, v) in http.headers {
        request.header(&h, v);
    }
    request
        .body(Body {
            content_length: Some(0),
            conn: Conn::new(0),
            read_buf: http.body,
        })
        .expect("should be able to create a body")
}

fn reader_to_request<R: Read>(mut c: R) -> Result<Request<Body>, Error> {
    let mut headers: Vec<httparse::Header> = vec![httparse::EMPTY_HEADER; MAX_HEADERS];
    let mut buf: Vec<u8> = Vec::new();

    let mut req = httparse::Request::new(&mut headers);
    c.read_to_end(&mut buf)?;
    let result = req.parse(&buf)?;
    if result.is_partial() {
        return Err(EmblyError::InvalidHttpRequest.into());
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
        conn: Conn::new(0),
        content_length: None,
        read_buf: buf[result.unwrap()..].to_vec(),
    })?)
}

// Will be used, currently just used for tests
#[allow(dead_code)]
fn reader_to_response<R: Read>(mut c: R) -> Result<Response<Body>, Error> {
    let mut headers: Vec<httparse::Header> = vec![httparse::EMPTY_HEADER; MAX_HEADERS];
    let mut buf: Vec<u8> = Vec::new();

    let mut res = httparse::Response::new(&mut headers);
    c.read_to_end(&mut buf)?;
    let result = res.parse(&buf)?;
    if result.is_partial() {
        return Err(EmblyError::InvalidHttpRequest.into());
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
        content_length: None,
        conn: Conn::new(0),
        read_buf: buf[result.unwrap()..].to_vec(),
    })?)
}

/// Run an http handler Function
///
/// ```no_run
/// use embly::http::{Body, Request, ResponseWriter};
/// use std::io::Write;
/// use embly::Error;
///
/// async fn execute (_req: Request<Body>, w: &mut ResponseWriter) -> Result<(), Error> {
///     w.write("hello world\n".as_bytes())?;
///     w.status("200")?;
///     w.header("Content-Length", "12")?;
///     w.header("Content-Type", "text/plain")?;
///     Ok(())
/// }
/// async fn catch_error(req: Request<Body>, mut w: ResponseWriter) {
///     match execute(req, &mut w).await {
///         Ok(_) => {}
///         Err(err) => {
///             w.status("500").unwrap();
///             w.write(format!("{}", err).as_bytes()).unwrap();
///         },
///     };
/// }
///
/// fn main() {
///     ::embly::http::run(catch_error);
/// }
///
pub fn run<F>(to_run: fn(Request<Body>, ResponseWriter) -> F)
where
    F: Future<Output = ()> + 'static,
{
    let function_id = 1;
    let mut c = Conn::new(function_id);
    let r = build_request_from_comm(&mut c).expect("http request should be valid");
    let mut resp = ResponseWriter::new(Body {
        content_length: None,
        conn: c.clone(),
        read_buf: Vec::new(),
    });
    task::Task::spawn(Box::pin(to_run(r, resp.clone())));
    resp.function_returned = true;
    resp.flush_response().expect("should be able to flush");
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn basic_parse() {
        let b = "hello";
        let e = reader_to_request(b.as_bytes())
            .err()
            .expect("there should be an error");
        match e
            .as_fail()
            .downcast_ref()
            .expect("the wrong type of error was returned")
        {
            EmblyError::InvalidHttpRequest => (),
            _ => panic!("the wrong error was returned"),
        }
    }

    #[test]
    fn simple_valid_request() -> Result<(), Error> {
        let b = "GET /c HTTP/1.1\r\nHost: f\r\n\r\n".as_bytes();
        reader_to_request(b)?;
        Ok(())
    }

    #[test]
    fn post_request() -> Result<(), Error> {
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

    // fn test_response_writer() -> ResponseWriter {
    //     ResponseWriter::new(Body {
    //         conn: Conn { id: 0 },
    //         read_buf: Vec::new(),
    //     })
    // }

    // #[test]
    // fn basic_response() -> Result<(), Error> {
    //     let mut w = test_response_writer();

    //     w.write_all(b"hello\n")?;
    //     w.status(401)?;
    //     w.header("Content-Type", "text/plain")?;

    //     w.function_returned = true;
    //     w.flush_response()?;
    //     let res = reader_to_response(w.body.conn)?;
    //     assert_eq!("6", res.headers().get("Content-Length").unwrap());
    //     assert_eq!(401, res.status());
    //     assert_eq!("text/plain", res.headers().get("Content-Type").unwrap());

    //     Ok(())
    // }

    // #[test]
    // fn no_status_response() -> Result<(), Error> {
    //     let mut w = test_response_writer();

    //     w.header("Content-Type", "text/plain")?;
    //     w.write_all(b"hello\n")?;
    //     w.function_returned = true;
    //     w.flush_response()?;

    //     let res = reader_to_response(w.body.conn)?;
    //     assert_eq!(200, res.status());
    //     Ok(())
    // }
}
