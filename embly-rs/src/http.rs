use crate::error::{Error, Result};
use crate::Comm;
use http;
use http::header::{HeaderName, HeaderValue};
use http::response::Parts;
use http::status::StatusCode;
use http::HttpTryFrom;
use http::{Request as HRequest, Response as HResponse};
use httparse;
use std::io;
use std::io::Read;
use std::io::Write;

pub type Request<T> = HRequest<T>;
pub type Response<T> = HResponse<T>;

pub struct Body {
    read_buf: Vec<u8>,
    comm: Comm,
}

pub struct ResponseWriter {
    body: Body,
    parts: Parts,
    write_buf: Vec<u8>,
    headers_writen: bool,
}

impl ResponseWriter {
    fn new(body: Body) -> Self {
        let (p, _) = Response::new(()).into_parts();
        Self {
            body: body,
            headers_writen: false,
            parts: p,
            write_buf: Vec::new(),
        }
    }
    fn write_headers(&mut self) -> Vec<u8> {
        let mut dst: Vec<u8> = Vec::new();
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
            // todo: content-length, chunked, etc....
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
    pub fn header<K, V>(&mut self, key: K, value: V) -> Result<()>
    where
        HeaderName: HttpTryFrom<K>,
        HeaderValue: HttpTryFrom<V>,
    {
        match HeaderName::try_from(key) {
            Ok(key) => match HeaderValue::try_from(value) {
                Ok(value) => {
                    self.parts.headers.append(key, value);
                    Ok(())
                }
                Err(e) => Err(Error::Http(e.into())),
            },
            Err(e) => Err(Error::Http(e.into())),
        }
    }
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
        if !self.headers_writen {
            let dst = &self.write_headers();
            self.body.write_all(dst)?;
        }
        self.body.write_all(&self.write_buf)?;
        self.write_buf.clear(); // is this ok? does this retain capacity?
        Ok(())
    }
}

#[inline]
fn extend(dst: &mut Vec<u8>, data: &[u8]) {
    dst.extend_from_slice(data);
}

impl io::Read for Body {
    fn read(&mut self, buf: &mut [u8]) -> io::Result<usize> {
        if self.read_buf.len() > 0 {
            let ln = (&self.read_buf[..]).read(buf)?;
            self.read_buf.truncate(self.read_buf.len() - ln);
            Ok(ln)
        } else {
            self.comm.read(buf)
        }
    }
}

impl io::Write for Body {
    fn write(&mut self, buf: &[u8]) -> io::Result<usize> {
        // todo: we're not just writing back to comm here
        // when we write to this message we're writing the
        // body so we need to be sure
        self.comm.write(buf)
    }
    fn flush(&mut self) -> io::Result<()> {
        // this might start a chunked response
        Ok(())
    }
}

fn build_request_from_comm(c: Comm) -> Result<Request<Body>> {
    c.wait();
    reader_to_request(c)
}

impl Body {}

// https://github.com/hyperium/hyper/blob/da9b0319ef8d85662f66ac3bea74036b3dd3744e/src/proto/h1/role.rs#L18
const MAX_HEADERS: usize = 100;
const AVERAGE_HEADER_SIZE: usize = 30;

fn reader_to_request<R: Read>(mut c: R) -> Result<Request<Body>> {
    let mut headers: Vec<httparse::Header> = vec![httparse::EMPTY_HEADER; MAX_HEADERS];
    let mut buf: Vec<u8> = Vec::new();

    let mut req = httparse::Request::new(&mut headers);
    c.read_to_end(&mut buf)?;
    println!("{:?}", buf);
    let res = req.parse(&buf)?;
    if res.is_partial() {
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
        if h.name.len() == 0 && h.value.len() == 0 {
            break;
        }
        request.header(h.name, h.value);
    }
    Ok(request.body(Body {
        comm: Comm { id: 0 },
        read_buf: buf[res.unwrap()..].to_vec(),
    })?)
}

pub fn run(to_run: fn(Request<Body>, &mut ResponseWriter) -> Result<()>) -> Result<()> {
    println!("running http func");

    let comm_id = 1;
    let c = Comm { id: comm_id };
    let r = build_request_from_comm(c)?;
    let mut resp = ResponseWriter::new(Body {
        comm: Comm { id: comm_id },
        read_buf: Vec::new(),
    });
    to_run(r, &mut resp)?;
    resp.flush()?;
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn basic_parse() {
        let b = "hello".as_bytes();
        let e = reader_to_request(b).err().unwrap();
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

field1=value1&field2=value2"
            .as_bytes();
        let mut request = reader_to_request(b)?;
        let body = request.body_mut();

        let mut b: Vec<u8> = Vec::new();
        body.read_to_end(&mut b)?;
        assert_eq!(b, "field1=value1&field2=value2".as_bytes());
        Ok(())
    }

}
