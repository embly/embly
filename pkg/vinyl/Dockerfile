FROM golang:1.12.9-stretch


RUN apt-get update && apt-get install -y build-essential
RUN go get golang.org/x/net/http2
WORKDIR /go/src/github.com/embly/vinyl/go-client/
