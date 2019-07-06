FROM docker:18.06.1 as docker
FROM golang:1.12-stretch

RUN apt-get update && apt-get install -y wget openssl dc ca-certificates git
RUN wget https://github.com/docker/compose/releases/download/1.23.0-rc2/docker-compose-Linux-x86_64 \
    && chmod +x docker-compose-Linux-x86_64 \
    && mv docker-compose-Linux-x86_64 /usr/local/bin/docker-compose
COPY --from=docker /usr/local/bin/docker /usr/local/bin/docker

WORKDIR /go/src/github.com/maxmcd/embly/

# RUN apk add git libc-dev gcc

RUN go get github.com/codegangsta/gin \
    && go get bitbucket.org/liamstask/goose/cmd/goose \
    && go get github.com/volatiletech/sqlboiler \
    && go get github.com/volatiletech/sqlboiler/drivers/sqlboiler-psql

RUN apt-get install -y \
    bash \
    curl \
    vim \
    tree \
    postgresql-client \
    sudo \
    `# end`

ARG EMBLY_USER

RUN adduser --disabled-password --gecos '' embly
RUN adduser embly sudo
RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers
RUN groupmod -g ${EMBLY_USER} embly
USER embly

COPY .bashrc /home/embly/.bashrc

ENV GO111MODULE=on

CMD bash