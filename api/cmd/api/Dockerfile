FROM golang:1.12-alpine3.10

RUN apk add git bash curl wget vim gcc libc-dev

RUN go get github.com/codegangsta/gin \
    && go get bitbucket.org/liamstask/goose/cmd/goose \
    && go get github.com/volatiletech/sqlboiler \
    && go get github.com/volatiletech/sqlboiler/drivers/sqlboiler-psql

ENV GO111MODULE=on
ENV PS1="\[\033[1;31m\]embly \[\033[1;37m\]\$(date +%M%S) \u: \w \n$ "

WORKDIR /go/src/github.com/maxmcd/embly/




# FROM alpine:edge
# WORKDIR /opt/
# COPY --from=0 /go/src/github.com/maxmcd/embly/api/cmd/api/api .
# CMD /opt/api

