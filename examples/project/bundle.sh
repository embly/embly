#! /usr/bin/env bash
set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

export RUST_BACKTRACE=1

cd ../../

cd cmd/embly

go install

cd ../../examples/project

embly bundle
