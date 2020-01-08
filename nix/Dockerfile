FROM docker:18.06.1 as docker

FROM debian:stretch-slim as ld
RUN apt-get update && apt-get install -y binutils patchelf
COPY ./embly /bin/embly
COPY ./embly-wrapper /bin/embly-wrapper
COPY ./lucetc /bin/lucetc
RUN patchelf --set-interpreter /lib64/ld-linux-x86-64.so.2 /bin/embly \
    && patchelf --set-interpreter /lib64/ld-linux-x86-64.so.2 /bin/embly-wrapper \
    && patchelf --set-interpreter /lib64/ld-linux-x86-64.so.2 /bin/lucetc

FROM gcr.io/distroless/cc:debug

COPY --from=docker /usr/local/bin/docker /usr/local/bin/docker
COPY --from=ld /bin/embly /bin/embly
COPY --from=ld /bin/embly-wrapper /bin/embly-wrapper
COPY --from=ld /bin/lucetc /bin/lucetc
COPY --from=ld /usr/bin/x86_64-linux-gnu-ld.bfd /bin/ld
COPY --from=ld \
    /usr/lib/x86_64-linux-gnu/libbfd-2.28-system.so \
    /usr/lib/x86_64-linux-gnu/
COPY --from=ld \
    /lib/x86_64-linux-gnu/libz.so.1 \
    /lib/x86_64-linux-gnu/libz.so.1.2.8 \
    /lib/x86_64-linux-gnu/libdl-2.24.so \
    /lib/x86_64-linux-gnu/libc-2.24.so \
    /lib/x86_64-linux-gnu/libdl.so.2 \
    /lib/x86_64-linux-gnu/libc.so.6 \
    /lib/x86_64-linux-gnu/

WORKDIR /app
