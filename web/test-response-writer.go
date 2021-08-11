package web

import (
	"bytes"
	"fmt"
	"net/http"
)

type TestResponseWriter struct {
	headerWritten bool
	header        http.Header
	buf           *bytes.Buffer
}

func NewTestResponseWriter() *TestResponseWriter {
	return &TestResponseWriter{
		headerWritten: false,
		header:        http.Header(map[string][]string{}),
		buf:           bytes.NewBuffer([]byte{}),
	}
}

func (w *TestResponseWriter) Header() http.Header {
	return w.header
}

func (w *TestResponseWriter) Write(bs []byte) (int, error) {
	err := w.writeHeader(http.StatusOK)
	if err != nil {
		return 0, nil
	}
	return w.buf.Write(bs)
}

func (w *TestResponseWriter) WriteHeader(code int) {
	w.writeHeader(code)
}

func (w *TestResponseWriter) writeHeader(code int) error {
	if w.headerWritten {
		return nil
	}
	w.headerWritten = true
	_, err := fmt.Fprintf(w.buf, "HTTP/1.1 %d %s\r\n", code, http.StatusText(code))
	if err != nil {
		return err
	}
	err = w.header.Write(w.buf)
	if err != nil {
		return err
	}
	_, err = w.buf.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	return nil
}
