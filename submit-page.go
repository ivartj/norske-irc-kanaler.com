package main

import (
	"net/http"
)

type submitErrorExcluded string
type submitErrorInvalid string
type submitErrorAlreadyIn string
type submitErrorApproval string

func (err submitErrorExcluded) Error() string { return string(err) }
func (err submitErrorInvalid) Error() string { return string(err) }
func (err submitErrorAlreadyIn) Error() string { return string(err) }
func (err submitErrorApproval) Error() string { return string(err) }

type submitOk string
func (err submitOk) Error() string { return string(err) }

func submitChannel(page *page, name, server, weblink, description string) error {

	name, server = channelAddressCanonical(name, server)

	if weblink == "" {
		weblink = channelSuggestWebLink(name, server)
	}

	err := channelAddressValidate(name, server)
	if err != nil {
		return submitErrorInvalid(err.Error())
	}

	isExcluded, excludeReason, err := dbIsChannelExcluded(page, name, server)
	if err != nil {
		panic(err)
	}

	if isExcluded {
		return submitErrorExcluded(excludeReason)
	}

	ch, _ := dbGetChannel(page, name, server)
	if ch != nil {
		return submitErrorAlreadyIn("Takk. Bidraget har allerede blitt sendt inn.")
	}

	err = dbAddChannel(page, name, server, weblink, description, !page.main.conf.Approval())
	if err != nil {
		panic(err)
	}

	if page.main.conf.Approval() {
		return submitErrorApproval("Takk for forslaget! Forslaget vil publiseres etter godkjenning av administrator.")
	} else {
		return submitOk("Takk for bidraget! Forslaget er publisert.")
	}
}

func submitPage(page *page, req *http.Request) {

	page.SetField("page-title", "Legg til chatterom")

	name := req.FormValue("name")
	network := req.FormValue("network")
	weblink := req.FormValue("weblink")
	description := req.FormValue("description")

	page.SetFieldMap(map[string]interface{}{
		"submit-name" : name,
		"submit-network" : network,
		"submit-weblink" : weblink,
		"submit-description" : description,

		"remove-form" : false,
	})

	if req.Method == "POST" {
		page.SetField("remove-form", true)

		err := submitChannel(page, name, network, weblink, description)
		page.AddMessage(err.Error())
		switch err.(type) {
		case submitErrorInvalid:
			page.SetField("remove-form", false)
		}
	}

	page.ExecuteTemplate("submit")
}

