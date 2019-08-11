package proxy

// stolen from https://golang.org/src/net/http/httputil/dump.go?s=5638:5700#L181

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func valueOrDefault(value, def string) string {
	if value != "" {
		return value
	}
	return def
}

var reqWriteExcludeHeaderDump = map[string]bool{
	"Host":              true, // not in Header map anyway
	"Transfer-Encoding": true,
	"Trailer":           true,
}

// DumpRequest dumps the http request headers
func DumpRequest(req *http.Request) (out []byte, err error) {
	var b bytes.Buffer

	reqURI := req.RequestURI
	if reqURI == "" {
		reqURI = req.URL.RequestURI()
	}

	fmt.Fprintf(&b, "%s %s HTTP/%d.%d\r\n", valueOrDefault(req.Method, "GET"),
		reqURI, req.ProtoMajor, req.ProtoMinor)

	absRequestURI := strings.HasPrefix(req.RequestURI, "http://") || strings.HasPrefix(req.RequestURI, "https://")
	if !absRequestURI {
		host := req.Host
		if host == "" && req.URL != nil {
			host = req.URL.Host
		}
		if host != "" {
			fmt.Fprintf(&b, "Host: %s\r\n", host)
		}
	}

	if len(req.TransferEncoding) > 0 {
		fmt.Fprintf(&b, "Transfer-Encoding: %s\r\n", strings.Join(req.TransferEncoding, ","))
	}
	if req.Close {
		fmt.Fprintf(&b, "Connection: close\r\n")
	}

	if err = req.Header.WriteSubset(&b, reqWriteExcludeHeaderDump); err != nil {
		return
	}

	io.WriteString(&b, "\r\n")
	out = b.Bytes()
	return
}
