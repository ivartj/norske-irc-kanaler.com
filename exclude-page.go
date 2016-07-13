package main

import (
	"net/http"
)

func excludePage(page *page, req *http.Request) {

	page.SetField("page-title", "Kanalekskludering")

	if req.Method == "GET" && req.FormValue("delete") == "yes" && req.FormValue("nonce") == page.main.auth.Nonce() {
		err := dbDeleteExclusion(
			page,
			req.FormValue("name"),
			req.FormValue("network"),
		)
		if err != nil {
			page.Fatalf("Failed to remove exclusion: %s", err.Error())
		}
		page.AddMessage("Ekskluderingen er fjernet.")
	}

	name, network, reason := "", "", ""

	if req.Method == "POST" {

		name = req.FormValue("name")
		network = req.FormValue("network")
		reason = req.FormValue("exclude-reason")
		name, network = channelAddressCanonical(name, network)
		err := channelAddressValidate(name, network)

		if err != nil {
			page.AddMessage("Ikke en gyldig kanal: %s", err.Error())
		} else {

			err = dbAddExclusion(page, name, network, reason)
			if err != nil {
				page.Fatalf("Failed to add exclusion: %s", err.Error())
			}

			page.AddMessage("Ekskludering lagt inn")
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

