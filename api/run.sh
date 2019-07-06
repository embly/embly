#!/usr/bin/env bash

cd "$(dirname ${BASH_SOURCE[0]})"
set -Euxo pipefail
# todo: check that this is executing in the embly environment

ATTEMPT=0
handle_close() {
    if [ $ATTEMPT -eq 0 ]; then
        ATTEMPT=1
        echo "Shutdown attempt. Try twice in quick succession if you would like to kill the process."
    else
        echo "Already tried to shutdown. Killing."
        exit 0
    fi
}
trap handle_close SIGINT SIGTERM

export GOOS=linux
export GOARCH=amd64

echo "Starting build and run loop:"
while :
do
    sleep 1 &
    wait $!
    ATTEMPT=0

    cd ./cmd/api/
    go build
    cd ../../
    cd ./cmd/rustcompile/
    go build
    cd ../../
    sudo docker-compose build
    sudo docker-compose up -d
    sudo docker-compose logs -f
done
