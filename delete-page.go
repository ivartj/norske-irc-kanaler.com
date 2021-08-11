package main

import (
	"net/http"
)

func deletePage(page *page, req *http.Request) {
	name := req.URL.Query().Get("name")
	network := req.URL.Query().Get("network")

	if page.main.auth.Nonce() != req.FormValue("nonce") {
		page.AddMessage("Nonce-mismatch.")
		page.ExecuteTemplate("message")
		return
	}

	err := dbDeleteChannel(page, name, network)
	if err != nil {
		page.Fatalf("Failed to delete channel: %s", err.Error())
	}
	page.AddMessage("%s@%s har blitt slettet.", name, network)
	page.ExecuteTemplate("delete")
}
