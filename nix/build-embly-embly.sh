#!/usr/bin/env bash
set -Eeuxo pipefail
cd "$(dirname ${BASH_SOURCE[0]})"

export DOCKER_BUILDKIT=1

./build-docker.sh

# ./run-docker.sh --run "make build_embly"
EMBLY_LOCATION=$(./run-docker.sh --run "which embly" | tr -d '\r')

export NIX_DOCKER_RUN_ARGS="-d" 
CONTAINER_ID=$(./run-docker.sh --run "sleep 5")

docker cp $CONTAINER_ID:/root/.cargo/bin/embly-wrapper ./embly-wrapper
docker cp $CONTAINER_ID:$EMBLY_LOCATION ./embly
docker cp $CONTAINER_ID:/root/.cargo/bin/lucetc ./lucetc

# maybe?
# https://blog.filippo.io/shrink-your-go-binaries-with-this-one-weird-trick/
strip embly-wrapper lucetc embly

docker kill $CONTAINER_ID
docker rm $CONTAINER_ID

docker build -f ./Dockerfile -t embly/embly .

rm embly-wrapper lucetc embly
