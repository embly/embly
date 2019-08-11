package proxy

import (
	"bytes"
	"log"
	"net/http/httptest"
	"testing"
)

func TestDumpRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "http://embly.org", bytes.NewBuffer([]byte("")))
	b, err := DumpRequest(req)
	if err != nil {
		t.Error(err)
	}
	log.Println(string(b))
}
