#!/usr/bin/env bash

set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

cd ./control
export EMBLY_USER=$(id -u)
docker network create embly || true
docker-compose down 
docker-compose run app