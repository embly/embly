#!/usr/bin/env bash
set -ex 
cd "$(dirname ${BASH_SOURCE[0]})"

docker build -f ./nix.Dockerfile -t embly-nix-build ..
