package main

import (
	"testing"
	"net/http"
	"net/url"
	"strings"
	"fmt"
)

func TestLoginPage(t *testing.T) {
	ctx := mainNewContext(testConf)
	rw := testNewResponseWriter()
	req, err := http.NewRequest("POST", "/login", strings.NewReader(""))
	if err != nil {
		t.Fatalf("Failed to create test request.\n")
	}
	req.Form = url.Values(map[string][]string{
		"password" : []string{ testConf.Password() },
	})

	ctx.site.ServeHTTP(rw, req)
	resp, err := rw.GetResponse(req)
	if err != nil {
		t.Fatalf("Failed to parse response to login: %s.\n", err.Error())
	}

	cookies := map[string]string{}
	for _, v := range resp.Cookies() {
		cookies[v.Name] = v.Value
	}

	sessionId, ok := cookies["session-id"]
	if !ok {
		t.Fatalf("No session ID cookie.\n")
	}

	fmt.Println(sessionId)
}

