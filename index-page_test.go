package main

import (
	"net/http"
	"strings"
	"testing"
)

func TestIndexPage(t *testing.T) {
	ctx := mainNewContext(testConf)
	req, err := http.NewRequest("GET", "/", strings.NewReader(""))
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %s.\n", err.Error())
	}
	ctx.site.ServeHTTP(testNewResponseWriter(), req)
}
