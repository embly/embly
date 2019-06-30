FROM docker:18.06.1 as docker
FROM golang:1.12-alpine3.10

ENV GLIBC 2.28-r0

RUN apk update && apk add --no-cache openssl ca-certificates curl libgcc && \
    curl -fsSL -o /etc/apk/keys/sgerrand.rsa.pub https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub && \
    curl -fsSL -o glibc-$GLIBC.apk https://github.com/sgerrand/alpine-pkg-glibc/releases/download/$GLIBC/glibc-$GLIBC.apk && \
    apk add --no-cache glibc-$GLIBC.apk \
    && ln -s /lib/libz.so.1 /usr/glibc-compat/lib/ \
    && ln -s /lib/libc.musl-x86_64.so.1 /usr/glibc-compat/lib \
    && ln -s /usr/lib/libgcc_s.so.1 /usr/glibc-compat/lib \
    && rm /etc/apk/keys/sgerrand.rsa.pub glibc-$GLIBC.apk \
    && wget https://github.com/docker/compose/releases/download/1.23.0-rc2/docker-compose-Linux-x86_64 \
    && chmod +x docker-compose-Linux-x86_64 \
    && mv docker-compose-Linux-x86_64 /usr/local/bin/docker-compose

COPY --from=docker /usr/local/bin/docker /usr/local/bin/docker

WORKDIR /go/src/github.com/maxmcd/embly/

RUN apk add git libc-dev gcc

RUN go get github.com/codegangsta/gin \
    && go get bitbucket.org/liamstask/goose/cmd/goose \
    && go get github.com/volatiletech/sqlboiler \
    && go get github.com/volatiletech/sqlboiler/drivers/sqlboiler-psql

RUN apk add \
    bash \
    curl \
    wget \
    vim \
    tree \
    `# end`

COPY .bashrc /root/.bashrc

ENV GO111MODULE=on

CMD bash