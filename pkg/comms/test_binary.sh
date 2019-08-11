set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

cd ../../embly-wrapper-rs
cargo build --release
ls -lah

cd ..
ROOT=$(pwd)

cd pkg/comms

export EMBLY_WRAPPER_BINARY_LOC="$ROOT/target/release/embly-wrapper-rs"
go test -v -benchmem -bench=. . 
