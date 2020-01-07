package httpproto

import (
	"bytes"
	"embly/pkg/tester"
	"fmt"
	"github.com/gogo/protobuf/proto"
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

func TestBasicDeserialize(te *testing.T) {
	t := tester.New(te)
	input := []byte{8, 1, 16, 1, 26, 1, 47, 42, 24, 10, 4, 72, 111, 115, 116, 18, 16, 10, 14, 108, 111, 99, 97, 108, 104, 111, 115, 116, 58, 56, 48, 56, 50, 42, 27, 10, 10, 85, 115, 101, 114, 45, 65, 103, 101, 110, 116, 18, 13, 10, 11, 99, 117, 114, 108, 47, 55, 46, 54, 55, 46, 48, 42, 15, 10, 6, 65, 99, 99, 101, 112, 116, 18, 5, 10, 3, 42, 47, 42}
	h := Http{}
	t.PanicOnErr(proto.Unmarshal(input, &h))
	fmt.Println(h)
}
