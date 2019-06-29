FROM golang:1.12-alpine3.10

RUN apk add git bash curl wget vim gcc libc-dev

RUN go get github.com/codegangsta/gin \
    && go get bitbucket.org/liamstask/goose/cmd/goose \
    && go get github.com/volatiletech/sqlboiler \
    && go get github.com/volatiletech/sqlboiler/drivers/sqlboiler-psql

ENV GO111MODULE=on
ENV PS1="\u: \w $ "

WORKDIR /go/src/github.com/maxmcd/embly/api/

RUN apk add git bash curl wget vim 

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN cd cmd/rustcompile && go build

FROM maxmcd/embly-compile-rust-wasm
WORKDIR /opt/app/
COPY --from=0 /go/src/github.com/maxmcd/embly/api/cmd/rustcompile .
CMD /opt/app/rustcmpile


