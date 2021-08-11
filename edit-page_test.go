package main

import (
	"net/http"
	"net/url"
	"testing"
)

func TestEditPage(t *testing.T) {

	ctx := mainNewContext(testConf)
	err := testSubmitChannel(ctx, "#test", "irc.example.com", "", "Lorem ipsum dolor sit amet.")
	if err != nil {
		t.Fatalf("Failed to submit channel to edit: %s.\n", err.Error())
	}

	sessionCookie, err := testLogin(ctx)
	if err != nil {
		t.Fatalf("Failed to log in: %s.\n", err.Error())
	}

	req, err := http.NewRequest("POST", "/edit", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %s.\n", err.Error())
	}

	req.AddCookie(sessionCookie)
	req.Form = url.Values(map[string][]string{
		"original-name":    []string{"#test"},
		"original-network": []string{"irc.example.com"},

		"name":        []string{"#fest"},
		"network":     []string{"irc.example.com"},
		"weblink":     []string{channelSuggestWebLink("#fest", "irc.example.com")},
		"description": []string{"Festkanalen."},
	})

	rw := testNewResponseWriter()

	ctx.site.ServeHTTP(rw, req)

	resp, err := rw.GetResponse(req)
	if err != nil {
		t.Fatalf("Failed to parse response to channel edit: %s.\n", err.Error())
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Status code after editing channel not 200 but %d.\n", resp.StatusCode)
	}

	_, err = dbGetChannel(ctx.db, "#fest", "irc.example.com")
	if err != nil {
		t.Fatalf("Failed to get edited channel: %s.\n", err.Error())
	}

}
