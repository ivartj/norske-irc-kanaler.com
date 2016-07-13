package main

import (
	"net/http"
)

func editPage(page *page, req *http.Request) {

	page.SetField("page-title", "Rediger kanal")

	originalName := ""
	originalNetwork := ""

	name := ""
	network := ""
	weblink := ""
	description := ""

	switch req.Method {
	case "GET":
		q := req.URL.Query()
		name = q.Get("name")
		network = q.Get("network")
		originalName = name
		originalNetwork = network
		ch, err := dbGetChannel(page, name, network)
		if err != nil {
			page.Fatalf("Failed to retrieve channel %s@%s from database: %s", originalName, originalNetwork, err.Error())
		}
		weblink = ch.Weblink()
		description = ch.Description()
	case "POST":
		name = req.FormValue("name")
		network = req.FormValue("network")
		originalName = req.FormValue("original-name")
		originalNetwork = req.FormValue("original-network")
		weblink = req.FormValue("weblink")
		description = req.FormValue("description")
		err := dbEditChannel(
			page,
			originalName,
			originalNetwork,
			name,
			network,
			weblink,
			description)
		if err != nil {
			page.Fatalf("Failed to edit channel: %s", err.Error())
		}
		page.AddMessage("Endring vellykket.")
	}

	page.SetFieldMap(map[string]interface{}{
		"edit-name" : name,
		"edit-network" : network,
		"edit-weblink" : weblink,
		"edit-description" : description,

		"edit-original-name" : originalName,
		"edit-original-network" : originalNetwork,
	})

	page.ExecuteTemplate("edit")

}

