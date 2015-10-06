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

func submitChannel(c *dbConn, name, server, weblink, description string) error {

	name, server = channelAddressCanonical(name, server)

	if weblink == "" {
		weblink = channelSuggestWebLink(name, server)
	}

	err := channelAddressValidate(name, server)
	if err != nil {
		return submitErrorInvalid(err.Error())
	}

	isExcluded, excludeReason, err := c.IsChannelExcluded(name, server)
	if err != nil {
		panic(err)
	}

	if isExcluded {
		return submitErrorExcluded(excludeReason)
	}

	ch, _ := c.GetChannel(name, server)
	if ch != nil {
		return submitErrorAlreadyIn("Takk. Bidraget har allerede blitt sendt inn.")
	}

	err = c.AddChannel(name, server, weblink, description, !conf.Approval)
	if err != nil {
		panic(err)
	}

	if conf.Approval {
		return submitErrorApproval("Takk for forslaget! Forslaget vil publiseres etter godkjenning av administrator.")
	} else {
		return submitOk("Takk for bidraget! Forslaget er publisert.")
	}
}

type serveSubmitContext struct {
	initialized bool
	Name string
	Network string
	Weblink string
	Description string
	Message string
	RemoveForm bool
	Excluded bool
}

func (ctx *serveContext) Submit() *serveSubmitContext {
	if ctx.submit.initialized {
		return &ctx.submit
	}

	ctx.submit.Name = ctx.req.FormValue("name")
	ctx.submit.Network = ctx.req.FormValue("network")
	ctx.submit.Weblink = ctx.req.FormValue("weblink")
	ctx.submit.Description = ctx.req.FormValue("description")

	ctx.submit.initialized = true

	return &ctx.submit
}

func (ctx *serveContext) serveSubmit(w http.ResponseWriter, req *http.Request) {
	ctx.setPageTitle("Legg til chatterom")

	if req.Method == "POST" {
		err := submitChannel(
			ctx.db,
			req.FormValue("name"),
			req.FormValue("network"),
			req.FormValue("weblink"),
			req.FormValue("description"),
		)
		switch err.(type) {
		case submitErrorAlreadyIn: ctx.submit.RemoveForm = true
		case submitErrorApproval: ctx.submit.RemoveForm = true
		case submitOk: ctx.submit.RemoveForm = true
		case submitErrorInvalid: ctx.submit.RemoveForm = false
		case submitErrorExcluded:
			ctx.submit.RemoveForm = true
			ctx.submit.Excluded = true
		}
		ctx.setMessage(err.Error())
	}

	err := ctx.executeTemplate(w, "submit", ctx)
	if err != nil {
		panic(err)
	}
}

