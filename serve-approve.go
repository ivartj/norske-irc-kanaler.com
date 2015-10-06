package main

import (
	"net/http"
	"fmt"
	"net/url"
)

type serveApproveContext struct {
	initialized bool
	Channels []channel
	ApproveName string
	ApproveNetwork string
	Message string
	MoreNext bool
	MorePrev bool
	PageNext int
	PagePrev int
}

func (ctx *serveContext) Approve() *serveApproveContext {

	if ctx.approve.initialized {
		return &ctx.approve
	}

	ctx.approve.ApproveName = ctx.req.URL.Query().Get("name")
	ctx.approve.ApproveNetwork = ctx.req.URL.Query().Get("network")
	if ctx.approve.ApproveName != "" && ctx.approve.ApproveNetwork != "" {
		err := ctx.db.ApproveChannel(ctx.approve.ApproveName, ctx.approve.ApproveNetwork)
		if err != nil {
			panic(err)
		}
		ctx.approve.Message = "Kanalen er godkjent!"
	}

	pagestr := ctx.req.URL.Query().Get("page")
	page := 1
	fmt.Sscanf(pagestr, "%d", &page)
	if page < 1 {
		page = 1
	}

	dbchs, err := ctx.db.GetUnapprovedChannels((page - 1) * 15, 15 + 1)
	if err != nil {
		panic(err)
	}
	if len(dbchs) > 15 {
		dbchs = dbchs[:15]
		ctx.approve.MoreNext = true
	}

	ctx.approve.MorePrev = page > 1
	ctx.approve.PageNext = page + 1
	ctx.approve.PagePrev = page - 1

	ctx.approve.Channels = dbchs

	ctx.approve.initialized = true
	return &ctx.approve
}

func (ctx *serveContext) serveApprove(w http.ResponseWriter, req *http.Request) {

	if loginAuth(req) == false {
		http.Redirect(w, req, "/login?redirect=" + url.QueryEscape(req.URL.Path + "?" + req.URL.RawQuery), 307)
		return
	}

	ctx.setPageTitle("Kanalgodkjenning")

	err := ctx.executeTemplate(w, "approve", ctx)
	if err != nil {
		panic(err)
	}
}

