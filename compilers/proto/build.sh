#! /usr/bin/env bash
set -Eeuxo pipefail
cd "$(dirname ${BASH_SOURCE[0]})"

docker-compose build app
docker --config ~/.docker-embly push embly/protoc

