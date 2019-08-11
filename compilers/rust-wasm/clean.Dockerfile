FROM intenal/embly-compile-rust-wasm:intermediate

FROM debian:stretch-slim
ENV CARGO_TARGET_DIR=/opt/target \
    RUSTUP_HOME=/usr/local/rustup \
    CARGO_HOME=/usr/local/cargo \
    RUST_VERSION=1.35.0 \
    PATH=/usr/local/cargo/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
COPY --from=0 /opt/basic-build/root/ /
WORKDIR /opt/
COPY ./basic-build /opt/basic-build
