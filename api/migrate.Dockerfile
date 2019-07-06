FROM golang:1.12-alpine3.9

WORKDIR /opt/

RUN apk add git libc-dev gcc

RUN go get github.com/codegangsta/gin \
    && go get bitbucket.org/liamstask/goose/cmd/goose \
    && go get github.com/volatiletech/sqlboiler \
    && go get github.com/volatiletech/sqlboiler/drivers/sqlboiler-psql

CMD cd pkg && goose up && sqlboiler psql