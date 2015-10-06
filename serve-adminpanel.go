package main

import (
	"net/http"
	"net/url"
)

type serveAdminPanelContext struct {
	initialized bool
	NumberForApproval int
	NumberExcluded int
}

func (ctx *serveContext) AdminPanel() *serveAdminPanelContext {

	if ctx.adminpanel.initialized {
		return &ctx.adminpanel
	}

	var err error
	ctx.adminpanel.NumberForApproval, err = ctx.db.GetNumberOfChannelsUnapproved()
	if err != nil {
		panic(err)
	}

	ctx.adminpanel.NumberExcluded, err = ctx.db.GetNumberOfChannelsExcluded()
	if err != nil {
		panic(err)
	}

	ctx.adminpanel.initialized = true
	return &ctx.adminpanel
}

func (ctx *serveContext) serveAdmin(w http.ResponseWriter, req *http.Request) {

	if loginAuth(req) == false {
		http.Redirect(w, req, "/login?redirect=" + url.QueryEscape(req.URL.Path + "?" + req.URL.RawQuery), 307)
		return
	}

	ctx.setPageTitle("Administratorpanel")

	err := ctx.executeTemplate(w, "admin", ctx)
	if err != nil {
		panic(err)
	}
}

