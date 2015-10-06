package main

import (
	"net/http"
	"net/url"
)

type serveExcludeContext struct {
	initialized bool
	ExcludeName string
	ExcludeNetwork string
	ExcludeReason string
	Message string
	Exclusions []*dbExclusion
}

func (ctx *serveContext) Exclude() *serveExcludeContext {

	if ctx.exclude.initialized {
		return &ctx.exclude
	}

	ctx.exclude.ExcludeName = ctx.req.URL.Query().Get("name")
	ctx.exclude.ExcludeNetwork = ctx.req.URL.Query().Get("network")

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
		err := ctx.db.AddExclusion(
			req.FormValue("name"),
			req.FormValue("network"),
			req.FormValue("exclude-reason"),
		)
		if err != nil {
			panic(err)
		}
		ctx.setMessage("Ekskludering lagt inn")
	}

	err := ctx.executeTemplate(w, "exclude", ctx)
	if err != nil {
		panic(err)
	}
}

