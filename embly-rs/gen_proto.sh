#! /usr/bin/env bash
set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

pb-rs -D -I ../pkg/core/httpproto/ -d src/http_proto ../pkg/core/httpproto/http.proto
