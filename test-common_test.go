package main

import (
	"net/http"
	"net/url"
	"fmt"
	"bytes"
	"bufio"
	"strings"
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

func testSubmitChannel(ctx *mainContext, name, network, weblink, description string) error {

	req, err := http.NewRequest("POST", "/submit", nil)
	if err != nil {
		return fmt.Errorf("Failed to create test request: %s", err.Error())
	}
	req.Form = url.Values(map[string][]string{
		"name" : []string{ name },
		"network" : []string{ network },
		"description" : []string{ description },
	})

	rw := testNewResponseWriter()

	ctx.site.ServeHTTP(rw, req)

	resp, err := rw.GetResponse(req)
	if err != nil {
		return fmt.Errorf("Failed to parse response: %s", err.Error())
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Response status code not 200, but %d", resp.StatusCode)
	}

	return nil
}

func testLogin(ctx *mainContext) (sessionCookie *http.Cookie, err error) {
	rw := testNewResponseWriter()
	req, err := http.NewRequest("POST", "/login", strings.NewReader(""))
	if err != nil {
		return nil, fmt.Errorf("Failed to create test request")
	}
	req.Form = url.Values(map[string][]string{
		"password" : []string{ ctx.conf.Password() },
	})

	ctx.site.ServeHTTP(rw, req)
	resp, err := rw.GetResponse(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse response to login: %s", err.Error())
	}

	cookies := map[string]*http.Cookie{}
	for _, v := range resp.Cookies() {
		cookies[v.Name] = v
	}

	sessionCookie, ok := cookies["session-id"]
	if !ok {
		return nil, fmt.Errorf("No session ID cookie")
	}

	return sessionCookie, nil
}

