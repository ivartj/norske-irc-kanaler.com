package main

import (
	"net/http"
	"net/url"
	"testing"
)

func TestExcludePage(t *testing.T) {

	ctx := mainNewContext(testConf)

	sessionCookie, err := testLogin(ctx)
	if err != nil {
		t.Fatalf("Failed to log in: %s.\n", err.Error())
	}

	rw := testNewResponseWriter()
	req, err := http.NewRequest("POST", "/exclude", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %s.\n", err.Error())
	}

	req.AddCookie(sessionCookie)
	req.Form = url.Values(map[string][]string{
		"name":           []string{"#test"},
		"network":        []string{"irc.example.com"},
		"exclude-reason": []string{"Too much trout-slapping."},
	})
	ctx.site.ServeHTTP(rw, req)

	resp, err := rw.GetResponse(req)
	if err != nil {
		t.Fatalf("Failed to parse response: %s.\n", err.Error())
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Status code when displaying exclusions not 200, but %d.\n", resp.StatusCode)
	}

}
