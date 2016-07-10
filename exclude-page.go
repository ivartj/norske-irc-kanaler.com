package main

import (
	"net/http"
	"github.com/ivartj/norske-irc-kanaler.com/web"
)

func excludePage(page web.Page, req *http.Request) {

	page.SetField("page-title", "Kanalekskludering")

	if req.Method == "GET" && req.FormValue("delete") == "yes" && req.FormValue("code") == loginSessionID {
		err := dbDeleteExclusion(
			page,
			req.FormValue("name"),
			req.FormValue("network"),
		)
		if err != nil {
			page.Fatalf("Failed to remove exclusion: %s", err.Error())
		}
		utilAddMessage(page, "Ekskluderingen er fjernet.")
	}

	name, network, reason := "", "", ""

	if req.Method == "POST" {

		name = req.FormValue("name")
		network = req.FormValue("network")
		reason = req.FormValue("exclude-reason")
		name, network = channelAddressCanonical(name, network)
		err := channelAddressValidate(name, network)

		if err != nil {
			utilAddMessage(page, "Ikke en gyldig kanal: %s", err.Error())
		} else {

			err = dbAddExclusion(page, name, network, reason)
			if err != nil {
				page.Fatalf("Failed to add exclusion: %s", err.Error())
			}

			utilAddMessage(page, "Ekskludering lagt inn")
		}
	}

	exclusions, err := dbGetExclusions(page)
	if err != nil {
		page.Fatalf("Failed to get exclusions: %s", err.Error())
	}

	page.SetFieldMap(map[string]interface{}{
		"exclude-name" : name,
		"exclude-network" : network,
		"exclude-reason" : reason,
		"exclusions" : exclusions,
	})

	page.ExecuteTemplate("exclude")
}

