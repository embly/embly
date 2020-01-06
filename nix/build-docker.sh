#!/usr/bin/env bash
set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

(cd .. && git archive --format=tar HEAD) | docker build -f ./nix/nix.Dockerfile -t embly-nix -
