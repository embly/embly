#! /usr/bin/env bash
set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

cd blog && ./build.sh
cd ../frontend && yarn run build

cd .. 
embly bundle
