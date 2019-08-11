package proxy

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"testing"
)

func TestDumpRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "http://embly.org", bytes.NewBuffer([]byte("")))
	b, err := DumpRequest(req)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(b))
}
