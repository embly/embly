set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

cargo +nightly build --target wasm32-wasi --release

ls -lah ../target/wasm32-wasi/release/embly.wasm
wasm-strip ../target/wasm32-wasi/release/embly.wasm
ls -lah ../target/wasm32-wasi/release/embly.wasm
if [ -x "$(command -v wasm2wat)" ]; then
    wasm2wat ../target/wasm32-wasi/release/embly.wasm > embly.wat
fi


