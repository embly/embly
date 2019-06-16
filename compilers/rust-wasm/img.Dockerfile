FROM r.j3ss.co/img 

WORKDIR /home/user/src

USER root
RUN echo "FROM maxmcd/embly-compile-rust-wasm" > /home/user/src/Dockerfile
USER user
