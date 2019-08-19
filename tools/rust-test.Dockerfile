FROM rust:1.36
WORKDIR /opt/

RUN PROTOC_ZIP=protoc-3.7.1-linux-x86_64.zip; \
    curl -OL https://github.com/google/protobuf/releases/download/v3.7.1/$PROTOC_ZIP; \
    unzip -o $PROTOC_ZIP -d /usr/local bin/protoc; \
    unzip -o $PROTOC_ZIP -d /usr/local include/*; \
    rm -f $PROTOC_ZIP

ENV CARGO_TARGET_DIR=/opt/target

RUN rustup component add clippy-preview

COPY Cargo.toml Cargo.toml
COPY embly-rs/Cargo.toml embly-rs/Cargo.toml
COPY embly-wrapper-rs/Cargo.toml embly-wrapper-rs/Cargo.toml

RUN mkdir -p \
    ./embly-rs/src \
    ./embly-wrapper-rs/src; \
    echo 'fn main(){ println!("hi") }' > ./embly-rs/src/lib.rs; \
    echo 'fn main(){ println!("hi") }' > ./embly-wrapper-rs/src/main.rs

RUN cargo fetch

COPY . .

RUN cargo test && cargo clippy \
    && cd /opt/examples/project && cargo clippy \
    && rm -rf /opt/target \
    && rm -rf /usr/local/cargo \
    && rm -rf /usr/local/rustup
# remove lots of artifacts so that the final CI upload is small
