package main

import (
	"bufio"
	"bytes"
	"log"
	"net/http"
	"testing"
)

func TestBasicHttp(t *testing.T) {
	b := bytes.NewBuffer([]byte(`HTTP/1.1 200 OK
content-length: 5
content-type: text/plain

hello`))
	resp, err := http.ReadResponse(bufio.NewReader(b), nil)
	if err != nil {
		t.Error(err)
	}
	log.Println(resp)
}
