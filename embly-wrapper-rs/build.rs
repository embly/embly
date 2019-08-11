extern crate protoc_rust;

use protoc_rust::Customize;

fn main() {
    protoc_rust::run(protoc_rust::Args {
        out_dir: "src/protos",
        input: &["../pkg/comms/proto/comms.proto"],
        includes: &["../pkg/comms/proto/"],
        customize: Customize {
            ..Default::default()
        },
    })
    .expect("protoc");
}
