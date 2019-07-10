#!/usr/bin/env bash

set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

sudo docker-compose up -d postgres
sleep 5
cd pkg
goose up
