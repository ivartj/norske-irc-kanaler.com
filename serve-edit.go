package main

import (
	"net/http"
	"fmt"
	"net/url"
)

type serveEditContext struct {
	initialized bool
	OriginalName string
	OriginalNetwork string
	Name string
	Network string
	Weblink string
	Description string
}

func (ctx *serveContext) Edit() *serveEditContext {
	if ctx.edit.initialized {
		return &ctx.edit
	}

	ctx.edit.initialized = true
	return &ctx.edit
}

func (ctx *serveContext) serveEdit(w http.ResponseWriter, req *http.Request) {
	if loginAuth(req) == false {
		http.Redirect(w, req, "/login?redirect=" + url.QueryEscape(req.URL.Path + "?" + req.URL.RawQuery), 307)
		return
	}

	ctx.setPageTitle("Rediger kanal")

	switch req.Method {
	case "GET":
		q := req.URL.Query()
		ctx.edit.OriginalName = q.Get("name")
		ctx.edit.Name = q.Get("name")
		ctx.edit.OriginalNetwork = q.Get("network")
		ctx.edit.Network = q.Get("network")
		ch, err := ctx.db.GetChannel(ctx.edit.OriginalName, ctx.edit.OriginalNetwork)
		if err != nil {
			panic(fmt.Errorf("Failed to retrieve channel %s@%s from database: %s", ctx.edit.OriginalName, ctx.edit.OriginalNetwork, err.Error()))
		}
		ctx.edit.Weblink = ch.Weblink()
		ctx.edit.Description = ch.Description()
	case "POST":
		ctx.edit.OriginalName = req.FormValue("name")
		ctx.edit.Name = req.FormValue("name")
		ctx.edit.OriginalNetwork = req.FormValue("network")
		ctx.edit.Network = req.FormValue("network")
		ctx.edit.Weblink = req.FormValue("weblink")
		ctx.edit.Description = req.FormValue("description")
		err := ctx.db.EditChannel(
			req.FormValue("original-name"),
			req.FormValue("original-network"),
			ctx.edit.Name,
			ctx.edit.Network,
			ctx.edit.Weblink,
			ctx.edit.Description)
		if err != nil {
			panic(err)
		}
		ctx.setMessage("Endring vellykket.")
	}

	err := ctx.executeTemplate(w, "edit", ctx)
	if err != nil {
		panic(err)
	}
}

