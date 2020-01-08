#! /usr/bin/env bash
set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"


export AWS_PROFILE=max
eval $(docker-machine env embly-run)
docker pull embly/app:latest
docker kill embly_run || true
docker rm embly_run || true
docker run --name=embly_run -d -p 8080:8082 embly/app
