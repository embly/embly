package httpproto

import (
	"bytes"
	"fmt"
	"log"
	"net/http/httptest"
	"testing"
)

func TestDumpRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "http://embly.org", bytes.NewBuffer([]byte("foo")))
	req.Header.Add("Content-Length", fmt.Sprint(req.ContentLength))
	request_proto, err := DumpRequest(req)
	if err != nil {
		t.Error(err)
	}
	log.Println(request_proto)
}
