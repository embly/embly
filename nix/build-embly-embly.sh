#!/usr/bin/env bash
set -Eeuxo pipefail
cd "$(dirname ${BASH_SOURCE[0]})"

export DOCKER_BUILDKIT=1

(cd .. && git archive --format=tar HEAD) | docker build -f ./nix/Dockerfile -t embly/embly -
