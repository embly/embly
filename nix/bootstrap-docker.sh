#!/usr/bin/env bash
set -ex
cd "$(dirname ${BASH_SOURCE[0]})"

docker volume create embly-nix-store
docker volume create embly-homedir

docker run -it -v embly-nix-store:/nix-dest embly-nix:latest sh -c "cp -r /nix/. /nix-dest"
docker run -it -v embly-homedir:/homedir-store embly-nix:latest sh -c "cp -r /root/. /homedir-store"
docker run -it -v embly-nix-store:/nix embly-nix:latest sh -c "nix-channel --add https://nixos.org/channels/nixpkgs-unstable nixpkgs"
docker run -it -v embly-nix-store:/nix embly-nix:latest sh -c "nix-channel --update"
docker run -it \
    -v embly-nix-store:/nix \
    -v embly-homedir:/root \
    embly-nix:latest \
sh -c "nix-env -iA nixpkgs.rustup nixpkgs.clang  \
&& nix-env -iA nixpkgs.cmake \
&& nix-env -iA nixpkgs.protobuf \
&& nix-env -iA nixpkgs.docker \
&& rustup toolchain add stable \
&& rustup toolchain add  nightly-2019-11-24 \
&& rustup target add wasm32-wasi --toolchain  nightly-2019-11-24 \
&& nix-shell -p bash --run 'cargo install lucetc || true'"
