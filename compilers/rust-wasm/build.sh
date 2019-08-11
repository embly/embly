#! /usr/bin/env bash
set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

export AWS_PROFILE=max

docker-compose build --no-cache app
# docker-compose build app

RUNNING_APP=$(docker-compose run -d app sleep 10000)

docker exec $RUNNING_APP bash -c "apt-get update && apt-get install -y strace"
docker cp ./basic-build $RUNNING_APP:/opt/basic-build

BUILD=$(cat <<END
cd /opt/basic-build
strace -f -v -o out.trace bash -c 'cargo +nightly build --target wasm32-wasi --release -Z unstable-options --out-dir ./out && wasm-strip ./out/*.wasm'
cat out.trace | grep "(\"/" | grep -v "/dev" > fileop.trace
cat fileop.trace | tr "," "\n" | grep -o '"/.*"' | sort | uniq > tocopy.trace
mkdir root && cat tocopy.trace | xargs -I{} cp --parents {} /opt/basic-build/root || true
END
)

docker exec $RUNNING_APP bash -c "$BUILD"

docker commit $RUNNING_APP intenal/embly-compile-rust-wasm:intermediate
docker kill $RUNNING_APP
docker rm $RUNNING_APP

docker-compose build clean