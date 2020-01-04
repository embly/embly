#! /usr/bin/env bash
set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

export RUST_BACKTRACE=1

cd ../../

cd cmd/embly

go install

cd ../../

cd embly-wrapper-rs
cargo install --force --path .

cd ../examples/mjpeg

embly run
