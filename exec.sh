#!/usr/bin/env bash

set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

docker exec -it $(docker ps --filter NAME=control_app -q) bash