package tester

var (
	ExternalDependency = map[string]string{
		"../lib/Cargo.toml": `
		`,
		"../lib/src/lib.rs": `
		`,
	}
	ExternalDependency2 = map[string]string{
		"../lib2/Cargo.toml": `
		`,
		"../lib2/src/lib.rs": `
		`,
	}
	EmblyFile = map[string]string{
		"./embly.hcl": `
function "auth" {
	runtime = "rust"
	path    = "./auth"
	sources = [
		"../lib",
	]
}
function "foo" {
	runtime = "rust"
	path    = "./foo"
	sources = [
		"../lib",
		"../lib2",
	]
}

files "frontend" {
	path              = "./frontend/build/"
	local_file_server = "http://localhost:3000"
}


files "blog" {
	path = "./blog/dist/"
}
`,
	}

	BasicRustProject = map[string]string{
		"./auth/Cargo.toml": `[package]
name = "hello"
version = "0.0.1"
edition = "2018"

[dependencies]
embly = "0.0.4"`,
		"./auth/src/main.rs": `
extern crate embly;
use embly::http::{Body, Request, ResponseWriter, run};
use embly::prelude::*;
use embly::Error;

fn execute(_req: Request<Body>, w: &mut ResponseWriter) -> Result<(), Error>{
	w.write_all(b"Hello World")?;
	Ok(())
}

fn main() -> Result<(), Error> {
	run(execute)
}`,
	}
	FooRustProject = map[string]string{
		"./foo/Cargo.toml": `[package]
name = "foo"
version = "0.0.1"
edition = "2018"`,
		"./foo/src/main.rs": `
fn main() {println!("hello")}
`,
	}
)
