package main

import (
	"testing"
	"net/http"
	"net/url"
)

func TestAdminPage(t *testing.T) {

	ctx := mainNewContext(testConf)

	rw := testNewResponseWriter()
	req, err := http.NewRequest("GET", "/admin", nil)
	if err != nil {
		t.Fatalf("Failed to create test request.\n")
	}
	ctx.site.ServeHTTP(rw, req)
	resp, err := rw.GetResponse(req)
	if err != nil {
		t.Fatalf("Failed to parse response: %s.\n", err.Error())
	}

	if resp.StatusCode < 300 || resp.StatusCode >= 400 {
		t.Fatalf("Unauthenticated request to /admin did not result in a redirect response code, but %s.\n", resp.Status)
	}

	rw = testNewResponseWriter()
	req, err = http.NewRequest("POST", "/login", nil)
	if err != nil {
		t.Fatalf("Failed to create test request.\n")
	}
	req.Form = url.Values(map[string][]string{
		"password" : []string { testConf.Password() },
	})

	ctx.site.ServeHTTP(rw, req)
	resp, err = rw.GetResponse(req)
	if err != nil {
		t.Fatalf("Failed to parse response: %s.\n", err.Error())
	}

	rw = testNewResponseWriter()
	req, err = http.NewRequest("GET", "/admin", nil)
	if err != nil {
		t.Fatalf("Failed to create test request.\n")
	}
	for _, v := range resp.Cookies() {
		req.AddCookie(v)
	}

	ctx.site.ServeHTTP(rw, req)
	resp, err = rw.GetResponse(req)
	if err != nil {
		t.Fatalf("Failed to parse response: %s.\n", err.Error())
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Status code on authenticated /admin request not 200 OK, but %s.\n", resp.Status)
	}
}

