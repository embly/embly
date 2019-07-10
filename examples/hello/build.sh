set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

cargo +nightly build --target wasm32-wasi --release

ls -lah ../../target/wasm32-wasi/release/hello.wasm
wasm-strip ../../target/wasm32-wasi/release/hello.wasm
ls -lah ../../target/wasm32-wasi/release/hello.wasm
if [ -x "$(command -v wasm2wat)" ]; then
    wasm2wat ../../target/wasm32-wasi/release/hello.wasm > hello.wat
fi


