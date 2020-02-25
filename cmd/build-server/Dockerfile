FROM golang:1.13-buster AS go

WORKDIR /embly

COPY go.mod go.sum ./
RUN go mod download
COPY pkg ./pkg
COPY cmd ./cmd

RUN cd cmd/build-server \
    && go build \
    && strip build-server

FROM nixos/nix as patchelf
COPY --from=go /embly/cmd/build-server/build-server /usr/bin/build-server
RUN export LIB="/nix/store/$(ls /nix/store | grep glibc)/lib" \
    && ls -lah $LIB \
    && nix-env -iA nixpkgs.patchelf \
    && patchelf --set-rpath "$LIB" /usr/bin/build-server \
    && patchelf --set-interpreter "$LIB/ld-linux-x86-64.so.2" /usr/bin/build-server

FROM nixos/nix
COPY --from=patchelf /usr/bin/build-server /usr/bin/build-server
