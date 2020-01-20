package httpproto

import (
	"io"

	protoutil "embly/pkg/protoutil"
)

type ReadWriter struct {
	ReadWriter io.ReadWriter
}

func (rw *ReadWriter) Write(b []byte) (ln int, err error) {
	if err = protoutil.WriteMessage(rw.ReadWriter, &Http{
		Body: b,
	}); err != nil {
		return
	}
	return len(b), nil
}

func (rw *ReadWriter) Next() (httpProto Http, err error) {
	err = protoutil.NextMessage(rw.ReadWriter, &httpProto)
	return
}
