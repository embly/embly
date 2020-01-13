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
//!curl

use crate::error::Error as EmblyError;
use crate::http_proto::httpproto::{HeaderList, Http};
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
#[derive(Debug, Default)]
pub struct Body {
    conn: Conn,
    content_length: Option<usize>,
    read_count: usize,
    read_buf: Vec<u8>,
}

struct Interior {
    body: Body,
    parts: Parts,
    write_buf: Vec<u8>,
}

impl Body {
    /// waits for all body bytes and returns them
    pub fn bytes(&mut self) -> Result<Vec<u8>, Error> {
        let mut out: Vec<u8> = self.read_buf.drain(..).collect();
        if self.content_length.is_none() || self.content_length.unwrap() == 0 {
            return Ok(out);
        }
        self.read_count = out.len();
        if self.read_count == self.content_length.unwrap() {
            Ok(out)
        } else {
            while self.read_count < self.content_length.unwrap() {
                let mut http = proto::next_message(&mut self.conn)?;
                self.read_count += http.body.len();
                out.append(&mut http.body);
            }
            Ok(out)
        }
    }
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
        let mut http_msg = Http::default();
        let mut interior = self.interior.lock().unwrap();

        if !self.headers_written {
            http_msg.status = interior.parts.status.as_u16() as i32;
            for (name, values) in interior.parts.headers.drain() {
                let mut list = HeaderList::default();
                for value in values {
                    list.header.push(value.to_str()?.to_string());
                }
                http_msg.headers.insert(name.as_str().to_string(), list);
            }
            self.headers_written = true;
        }
        if self.function_returned {
            http_msg.eof = true;
        }
        http_msg.body = interior.write_buf.drain(..).collect();
        proto::write_msg(&mut interior.body.conn, http_msg)?;
        Ok(())
    }
}

// impl io::Read for Body {
//     fn read(&mut self, buf: &mut [u8]) -> io::Result<usize> {
//         if !self.read_buf.is_empty() {
//             let ln = (&self.read_buf[..]).read(buf)?;
//             self.read_buf.drain(0..ln);
//             Ok(ln)
//         } else {
//             self.conn.read(buf)
//         }
//     }
// }

// impl io::Write for Body {
//     fn write(&mut self, buf: &[u8]) -> io::Result<usize> {
//         // todo: we're not just writing back to connection here
//         // when we write to this message we're writing the
//         // body so we need to be sure
//         self.conn.write(buf)
//     }
//     fn flush(&mut self) -> io::Result<()> {
//         Ok(())
//     }
// }

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

fn http_proto_to_request(http: Http) -> Request<Body> {
    let mut request = Request::builder();
    request.uri(http.uri);
    // hardcode a map?
    let mut body = Body::default();
    request.method(format!("{:?}", http.method).as_str());
    for (h, values) in http.headers {
        if h == "Content-Length" {
            let cl: usize = values.header[0]
                .parse()
                .expect("content length should be an int");
            body.content_length = Some(cl);
        }
        for v in values.header {
            request.header(&h, v);
        }
    }
    body.read_buf = http.body;
    request.body(body).expect("should be able to create a body")
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
    let mut body = Body::default();
    body.read_buf = buf[result.unwrap()..].to_vec();
    Ok(response.body(body)?)
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
    let mut c = Conn::new(1);
    let r = build_request_from_comm(&mut c).expect("http request should be valid");
    let mut body = Body::default();
    body.conn = c.clone();
    let mut resp = ResponseWriter::new(body);
    task::Task::spawn(Box::pin(to_run(r, resp.clone())));
    resp.function_returned = true;
    resp.flush_response().expect("should be able to flush");
}

/// Run an http handler Function, but return a 500 response if the handler returns an error
///
/// ```no_run
/// use embly::http::{Body, Request, ResponseWriter};
/// use std::io::Write;
/// use embly::Error;
///
/// async fn execute (_req: Request<Body>, mut w: ResponseWriter) -> Result<(), Error> {
///     w.write("hello world\n".as_bytes())?;
///     w.status("200")?;
///     w.header("Content-Length", "12")?;
///     w.header("Content-Type", "text/plain")?;
///     Ok(())
/// }
///
/// fn main() {
///     ::embly::http::run_catch_error(execute);
/// }
///
pub fn run_catch_error<F>(to_run: fn(Request<Body>, ResponseWriter) -> F)
where
    F: Future<Output = Result<(), Error>> + 'static,
{
    let mut c = Conn::new(1);
    let r = build_request_from_comm(&mut c).expect("http request should be valid");
    let mut body = Body::default();
    body.conn = c.clone();
    let mut resp = ResponseWriter::new(body);
    let user_resp = resp.clone();
    let mut error_resp = resp.clone();
    task::Task::spawn(Box::pin(async move {
        match to_run(r, user_resp).await {
            Ok(_) => {}
            Err(err) => {
                println!("got error: {}", err);
                error_resp.status(500).unwrap();
                error_resp.write(&format!("{}", err).as_bytes()).unwrap();
            }
        }
    }));
    resp.function_returned = true;
    resp.flush_response().expect("should be able to flush");
}

#[cfg(test)]
mod tests {
    // use super::* ;

    //     #[test]
    //     fn post_request() -> Result<(), Error> {
    //         let b = "POST /test HTTP/1.1
    // Host: foo.example
    // Content-Type: application/x-www-form-urlencoded
    // Content-Length: 27

    // field1=value1&field2=value2";
    //         let mut request = reader_to_request(b.as_bytes())?;
    //         let body = request.body_mut();
    //         let b: Vec<u8> = body.bytes()?;
    //         let values = "field1=value1&field2=value2";
    //         assert_eq!(b, values.as_bytes());
    //         Ok(())
    //     }

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
