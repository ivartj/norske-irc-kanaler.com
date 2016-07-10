package main

import (
	"net/http"
	"github.com/ivartj/norske-irc-kanaler.com/web"
)

func adminPage(page web.Page, req *http.Request) {
	page.SetField("page-title", "Administratorpanel")

	numUnapproved, err := dbGetNumberOfChannelsUnapproved(page)
	if err != nil {
		page.Fatalf("Failed to get number of unapproved channels: %s", err.Error())
	}

	numExcluded, err := dbGetNumberOfChannelsExcluded(page)
	if err != nil {
		page.Fatalf("Failed to get number of excluded channels: %s", err.Error())
	}

	page.SetFieldMap(map[string]interface{}{
		"number-for-approval" : numUnapproved,
		"number-excluded" : numExcluded,
	})

	page.ExecuteTemplate("admin")
}

