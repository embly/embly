package httpproto

// stolen from https://golang.org/src/net/http/httputil/dump.go?s=5638:5700#L181

import (
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
func DumpRequest(req *http.Request) (out Http, err error) {
	out.Headers = make(map[string]*HeaderList)
	reqURI := req.RequestURI
	if reqURI == "" {
		reqURI = req.URL.RequestURI()
	}
	out.Method = Http_Method(Http_Method_value[valueOrDefault(req.Method, "GET")])
	out.Uri = reqURI
	out.ProtoMajor = int32(req.ProtoMajor)
	out.ProtoMinor = int32(req.ProtoMinor)

	absRequestURI := strings.HasPrefix(req.RequestURI, "http://") || strings.HasPrefix(req.RequestURI, "https://")
	if !absRequestURI {
		host := req.Host
		if host == "" && req.URL != nil {
			host = req.URL.Host
		}
		if host != "" {
			out.Headers["Host"] = &HeaderList{Header: []string{host}}
		}
	}

	if len(req.TransferEncoding) > 0 {
		out.Headers["Transfer-Encoding"] = &HeaderList{Header: []string{
			strings.Join(req.TransferEncoding, ","),
		}}
	}
	if req.Close {
		out.Headers["Connection"] = &HeaderList{Header: []string{
			"close",
		}}
	}

	for k, vs := range req.Header {
		if reqWriteExcludeHeaderDump[k] {
			continue
		}
		out.Headers[k] = &HeaderList{Header: vs}
	}

	return
}
