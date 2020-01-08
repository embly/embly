#!/usr/bin/env bash
set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

export AWS_ACCESS_KEY_ID=$(cat ~/.aws/credentials | grep -A 3 "\[max\]" | grep aws_access_key_id | grep -o '[^ ]\+$')
export AWS_SECRET_ACCESS_KEY=$(cat ~/.aws/credentials | grep -A 3 "\[max\]" | grep aws_secret_access_key | grep -o '[^ ]\+$')

docker-compose build