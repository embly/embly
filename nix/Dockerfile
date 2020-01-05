FROM docker:18.06.1 as docker
FROM golang:1.13 as embly

WORKDIR /opt

COPY go.mod go.sum ./
RUN go mod download
COPY ./pkg/ ./pkg
COPY ./cmd/ ./cmd

RUN ls -lah && cd cmd/embly && go install

FROM rust:1.40-slim-stretch as lucetc
RUN apt-get update && apt-get install -y build-essential cmake python3
RUN cargo install --version=0.4.1 lucetc
RUN apt-get install -y binutils 

FROM rust:1.40-slim-stretch as wrapper
RUN apt-get update && apt-get install -y python3 protobuf-compiler cmake 
WORKDIR /opt
COPY embly-wrapper-rs/Cargo.toml embly-wrapper-rs/Cargo.toml
COPY pkg/core/proto pkg/core/proto
RUN mkdir -p \
    ./embly-wrapper-rs/src \
    && echo 'fn main(){ println!("hi") }' > ./embly-wrapper-rs/src/main.rs 
RUN cd embly-wrapper-rs && cargo fetch
COPY embly-wrapper-rs embly-wrapper-rs 
RUN cd embly-wrapper-rs && cargo build --release

FROM gcr.io/distroless/cc:debug

COPY --from=docker /usr/local/bin/docker /usr/local/bin/docker
COPY --from=embly /go/bin/embly /bin/embly
COPY --from=wrapper /opt/embly-wrapper-rs/target/release/embly-wrapper /bin/embly-wrapper
COPY --from=lucetc /usr/local/cargo/bin/lucetc /bin/lucetc
COPY --from=lucetc /usr/bin/x86_64-linux-gnu-ld.bfd /bin/ld
COPY --from=lucetc \
    /usr/lib/x86_64-linux-gnu/libbfd-2.28-system.so \
    /usr/lib/x86_64-linux-gnu/
COPY --from=lucetc \
    /lib/x86_64-linux-gnu/libz.so.1 \
    /lib/x86_64-linux-gnu/libz.so.1.2.8 \
    /lib/x86_64-linux-gnu/libdl-2.24.so \
    /lib/x86_64-linux-gnu/libc-2.24.so \
    /lib/x86_64-linux-gnu/libdl.so.2 \
    /lib/x86_64-linux-gnu/libc.so.6 \
    /lib/x86_64-linux-gnu/

WORKDIR /app
