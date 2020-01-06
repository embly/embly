#!/usr/bin/env bash
set -ex
cd "$(dirname ${BASH_SOURCE[0]})"

docker run -it \
    -v embly-nix-store:/nix \
    -v embly-homedir:/root \
    -v /var/run/docker.sock:/var/run/docker.sock \
    --workdir /opt \
    embly-nix:latest \
nix-shell /opt/nix/sccache-shell.nix "$@"
