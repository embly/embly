package proxy

import "net/http"

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	DumpRequest(r)
}
