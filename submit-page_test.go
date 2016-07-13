package main

import (
	"testing"
	"net/url"
	"net/http"
)

func TestSubmitPage(t *testing.T) {

	// TODO: Check that submit web form field names reflect the the form
	//       names read

	ctx := mainNewContext(testConf)

	req, err := http.NewRequest("POST", "/submit", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %s.\n", err.Error())
	}
	req.Form = url.Values(map[string][]string{
		"name" : []string{ "#test" },
		"network" : []string{ "irc.example.com" },
		"description" : []string{ "Lorem ipsum dolor sit amet." },
	})

	rw := testNewResponseWriter()

	ctx.site.ServeHTTP(rw, req)

	resp, err := rw.GetResponse(req)
	if err != nil {
		t.Fatalf("Failed to parse response: %s.\n", err.Error())
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Response status code not 200 OK, but: %s.\n", resp.Status)
	}

	ch, err := dbGetChannel(ctx.db, "#test", "irc.example.com")
	if err != nil {
		t.Fatalf("Failed to retrieve submitted channel: %s.\n", err.Error())
	}

	if ch.Name() != "#test" || ch.Network() != "irc.example.com" {
		t.Fatalf("Retrieved channel %s@%s not the submitted channel.\n", ch.Name(), ch.Network())
	}
}

