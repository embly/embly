FROM embly/embly as embly

FROM gcr.io/distroless/cc:debug
WORKDIR /opt
COPY --from=embly /bin/embly /bin/embly
COPY --from=embly /bin/embly-wrapper /bin/embly-wrapper
COPY out.tar.gz .
COPY embly.hcl embly.hcl

ENTRYPOINT
CMD ["embly", "run", "--host", "0.0.0.0", "out.tar.gz"]
