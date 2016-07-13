package main

import (
	"net/http"
	"fmt"
	"bytes"
	"bufio"
)

var testConf = &conf{confSet{
	WebsiteTitle: "Norske IRC-kanaler",
	WebsiteDescription: "Oversikt over norske IRC-kanaler.",
	DatabasePath: ":memory:",
	AssetsPath: "./assets",
	Password: "lutefisk",
}}

type testResponseWriter struct{
	headerWritten bool
	header http.Header
	buf *bytes.Buffer
}

func testNewResponseWriter() *testResponseWriter {
	return &testResponseWriter{
		headerWritten: false,
		header: http.Header(map[string][]string{}),
		buf: bytes.NewBuffer([]byte{}),
	}
}

func (w *testResponseWriter) Header() http.Header {
	return w.header
}

func (w *testResponseWriter) Write(bs []byte) (int, error) {
	err := w.writeHeader(http.StatusOK)
	if err != nil {
		return 0, nil
	}
	return w.buf.Write(bs)
}

func (w *testResponseWriter) WriteHeader(code int) {
	w.writeHeader(code)
}

func (w *testResponseWriter) writeHeader(code int) error {
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

func (w *testResponseWriter) GetResponse(req *http.Request) (*http.Response, error) {
	return http.ReadResponse(bufio.NewReader(w.buf), req)
}

