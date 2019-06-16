set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

goimports -w .

export GO111MODULE=auto

cd ../embly
./build.sh
cd ../local-runtime

export WASM_LOCATION="../target/wasm32-wasi/release/embly.wasm"
# export WASM_LOCATION="/Users/maxm/go/src/github.com/ajkavanagh/rust-mandelbrot/target/wasm32-wasi/release/mandelbrot.wasm"

# ls -lah $WASM_LOCATION

go run . $@
