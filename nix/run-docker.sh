#!/usr/bin/env bash
set -ex
cd "$(dirname ${BASH_SOURCE[0]})"
D="${NIX_DOCKER_RUN_ARGS}"

docker run -it $D \
    -v embly-nix-store:/nix \
    -v embly-homedir:/root \
    -v /var/run/docker.sock:/var/run/docker.sock \
    --workdir /opt \
    embly-nix:latest \
nix-shell /opt/nix/sccache-shell.nix "$@"
