package main

import (
	"testing"
)

func TestSubmitPage(t *testing.T) {

	// TODO: Check that submit web form field names reflect the the form
	//       names read

	ctx := mainNewContext(testConf)
	err := testSubmitChannel(ctx, "#test", "irc.example.com", "", "Lorem ipsum dolor sit amet.")
	if err != nil {
		t.Fatalf("Failed to submit channel: %s.\n", err.Error())
	}

	ch, err := dbGetChannel(ctx.db, "#test", "irc.example.com")
	if err != nil {
		t.Fatalf("Failed to retrieve submitted channel: %s.\n", err.Error())
	}

	if ch.Name() != "#test" || ch.Network() != "irc.example.com" {
		t.Fatalf("Retrieved channel %s@%s not the submitted channel.\n", ch.Name(), ch.Network())
	}
}

