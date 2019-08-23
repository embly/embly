#! /usr/bin/env bash
set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

cd ..

cd embly-wrapper-rs
cargo build

cd ..
cd pkg/comms

go generate
