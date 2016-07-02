package main

import (
	"net/http"
	"net/url"
	"fmt"
)

type serveExcludeContext struct {
	initialized bool
	ExcludeName string
	ExcludeNetwork string
	ExcludeReason string
	Exclusions []*dbExclusion
}

func (ctx *serveContext) Exclude() *serveExcludeContext {

	if ctx.exclude.initialized {
		return &ctx.exclude
	}

	var err error
	ctx.exclude.Exclusions, err = ctx.db.GetExclusions()
	if err != nil {
		panic(err)
	}

	ctx.exclude.initialized = true
	return &ctx.exclude
}

func (ctx *serveContext) serveExclude(w http.ResponseWriter, req *http.Request) {

	if loginAuth(req) == false {
		http.Redirect(w, req, "/login?redirect=" + url.QueryEscape(req.URL.Path + "?" + req.URL.RawQuery), 307)
		return
	}

	ctx.setPageTitle("Kanalekskludering")

	if req.Method == "GET" && req.FormValue("delete") == "yes" && req.FormValue("code") == loginSessionID {
		err := ctx.db.DeleteExclusion(
			req.FormValue("name"),
			req.FormValue("network"),
		)
		if err != nil {
			panic(err)
		}
		ctx.setMessage("Ekskluderingen er fjernet")
	}

	if req.Method == "POST" {

		name := ctx.req.URL.Query().Get("name")
		network := ctx.req.URL.Query().Get("network")
		reason := req.FormValue("exclude-reason")

		name, network = channelAddressCanonical(name, network)
		err := channelAddressValidate(name, network)

		if err != nil {
			ctx.setMessage(fmt.Sprintf("Ikke en gyldig kanaladresse: %s", err.Error()))
		} else {

			err = ctx.db.AddExclusion(name, network, reason)
			if err != nil {
				panic(err)
			}

			ctx.setMessage("Ekskludering lagt inn")
		}

		ctx.exclude.ExcludeName = name
		ctx.exclude.ExcludeNetwork = network
		ctx.exclude.ExcludeReason = reason
	}

	err := ctx.executeTemplate(w, "exclude", ctx)
	if err != nil {
		panic(err)
	}
}

