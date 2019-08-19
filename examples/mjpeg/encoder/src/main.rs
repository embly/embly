extern crate embly;
extern crate image;
extern crate num_complex;

use embly::error::Result;
use embly::http;
use embly::http::{Body, Flusher, Request, ResponseWriter};
use image::jpeg;
use std::io::Write;

use std::{thread, time};

fn execute(_req: Request<Body>, w: &mut ResponseWriter) -> Result<()> {
    w.status("200")?;

    let boundary = "ebf0e0ed5f2263de9bf04ab6833a706a4877044378e7f740865a263b1664";
    w.header(
        "Content-Type",
        format!("multipart/x-mixed-replace; boundary={}", boundary),
    )?;
    w.header("Connection", "close")?;

    w.flush_response()?;
    let one_second = time::Duration::new(0, 1_000_000_000);
    let start_time = time::SystemTime::now()
        .duration_since(time::SystemTime::UNIX_EPOCH)
        .unwrap()
        .as_secs();
    let mut index = 0;
    loop {
        w.write(format!("--{}\r\n", boundary).as_bytes())?;
        w.write(b"Content-Type: image/jpeg\r\n")?;
        w.write(b"Content-Length: ")?;
        let bytes = gen_dot(index)?;
        w.write(bytes.len().to_string().as_bytes())?;
        w.write(b"\r\n")?;

        w.write(format!("X-Starttime: {}\r\n", start_time).as_bytes())?;
        let now = time::SystemTime::now()
            .duration_since(time::SystemTime::UNIX_EPOCH)
            .unwrap()
            .as_secs();
        w.write(format!("X-Timestamp: {}\r\n", now).as_bytes())?;

        w.write(b"\r\n")?;
        w.write(&bytes)?;
        thread::sleep(one_second);
        w.flush_response()?;
        index += 1;
    }
}

fn main() -> Result<()> {
    http::run(execute)
}

fn gen_dot(index: u32) -> Result<Vec<u8>> {
    let imgx = 10;
    let imgy = 10;
    let mut imgbuf = image::ImageBuffer::new(imgx, imgy);
    let pixel = imgbuf.get_pixel_mut(index % 10, index / 10 % 10);
    *pixel = image::Rgb([255u8, 255u8, 255u8]);
    let mut buffer: Vec<u8> = Vec::new();
    jpeg::JPEGEncoder::new(&mut buffer).encode(&imgbuf, imgx, imgy, image::ColorType::RGB(8))?;

    // png::PNGEncoder::new(&mut w).encode(&imgbuf, imgx, imgy, image::ColorType::RGB(8))?;

    Ok(buffer)
}

#[allow(dead_code)]
fn gen_fractial<W: Write>(mut w: W) -> Result<()> {
    let imgx = 400;
    let imgy = 400;

    let scalex = 3.0 / imgx as f32;
    let scaley = 3.0 / imgy as f32;

    let mut imgbuf = image::ImageBuffer::new(imgx, imgy);
    for (x, y, pixel) in imgbuf.enumerate_pixels_mut() {
        let r = (0.3 * x as f32) as u8;
        let b = (0.3 * y as f32) as u8;
        *pixel = image::Rgb([r, 0, b]);
    }
    // A redundant loop to demonstrate reading image data
    for x in 0..imgx {
        for y in 0..imgy {
            let cx = y as f32 * scalex - 1.5;
            let cy = x as f32 * scaley - 1.5;

            let c = num_complex::Complex::new(-0.4, 0.6);
            let mut z = num_complex::Complex::new(cx, cy);

            let mut i = 0;
            while i < 255 && z.norm() <= 2.0 {
                z = z * z + c;
                i += 1;
            }

            let pixel = imgbuf.get_pixel_mut(x, y);
            let image::Rgb(data) = *pixel;
            *pixel = image::Rgb([data[0], i as u8, data[2]]);
        }
    }
    jpeg::JPEGEncoder::new(&mut w).encode(&imgbuf, imgx, imgy, image::ColorType::RGB(8))?;
    Ok(())
}
