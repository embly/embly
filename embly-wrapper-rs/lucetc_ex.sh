#! /usr/bin/env bash
set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

lucetc \
    --bindings ./bindings.json \
    --opt-level 2 \
    ../examples/project/embly_build/listener.wasm 