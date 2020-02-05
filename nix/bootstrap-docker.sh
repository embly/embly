#!/usr/bin/env bash
set -ex
cd "$(dirname ${BASH_SOURCE[0]})"

docker volume create embly-nix-store
docker volume create embly-homedir

docker run -it -v embly-nix-store:/nix-dest embly-nix:latest sh -c "cp -r /nix/. /nix-dest"
docker run -it -v embly-homedir:/homedir-store embly-nix:latest sh -c "cp -r /root/. /homedir-store && mkdir -p /homedir-store/target"
docker run -it -v embly-nix-store:/nix embly-nix:latest sh -c "nix-channel --add https://nixos.org/channels/nixpkgs-unstable nixpkgs"
docker run -it -v embly-nix-store:/nix embly-nix:latest sh -c "nix-channel --update"
docker run -it \
    -v embly-nix-store:/nix \
    -v embly-homedir:/root \
    embly-nix:latest \
sh -c "nix-env -iA nixpkgs.rustup nixpkgs.gcc  \
nixpkgs.cmake \
nixpkgs.protobuf \
nixpkgs.docker \
nixpkgs.python3 \
nixpkgs.zola \
&& rustup toolchain add stable \
&& rustup toolchain add  nightly-2019-11-24 \
&& rustup target add wasm32-wasi --toolchain  nightly-2019-11-24 \
&& nix-shell -p bash --run 'if [ \"lucetc 0.4.3\" != \"\$(lucetc --version)\" ]; then cargo install --force --version 0.4.3 lucetc; fi'"
