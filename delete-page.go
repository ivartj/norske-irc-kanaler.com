package main

import (
	"net/http"
	"github.com/ivartj/norske-irc-kanaler.com/web"
)

func deletePage(page web.Page, req *http.Request) {
	name := req.URL.Query().Get("name")
	network := req.URL.Query().Get("network")
	page.SetField("referer", req.Referer())
	err := dbDeleteChannel(page, name, network)
	if err != nil {
		page.Fatalf("Failed to delete channel: %s", err.Error())
	}
	utilAddMessage(page, "%s@%s har blitt slettet.", name, network)
	page.ExecuteTemplate("delete")
}

