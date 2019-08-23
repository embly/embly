FROM golang:1.12-stretch
WORKDIR /opt/

ENV GO111MODULE=on

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY . .
RUN go test -race $(go list ./... | grep -v models)
