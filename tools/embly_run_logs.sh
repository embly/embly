#! /usr/bin/env bash
set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

export AWS_PROFILE=max
eval $(docker-machine env embly-run)
docker logs -f embly_run
