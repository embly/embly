package httpproto

import "io"

import protoutil "embly/pkg/proto-util"

type Writer struct {
	Writer io.Writer
}

func (wr *Writer) Write(b []byte) (ln int, err error) {
	if err = protoutil.WriteMessage(wr.Writer, &Http{
		Body: b,
	}); err != nil {
		return
	}
	return len(b), nil
}
