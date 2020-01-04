#!/usr/bin/env bash
set -ex

docker run -it -v pesto_nix-store:/nix embly-nix-build:latest sh
