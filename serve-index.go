package main

import (
	"net/http"
	"fmt"
)

type indexChannel struct{
	Name string
	Server string
	WebLink string
	Description string
	Status string
	Error bool
}

// Component of serveContext
type serveIndexContext struct {
	initialized bool
	MoreNext, MorePrev bool
	PageNext, PagePrev int
	Channels []channel
}

func (ctx *serveContext) Index() *serveIndexContext {
	if ctx.index.initialized {
		return &ctx.index
	}

	page := 1
	pagestr := ctx.req.URL.Query().Get("page")
	if pagestr != "" {
		fmt.Sscanf(pagestr, "%d", &page)
		if page < 1 {
			page = 1
		}
	}

	// TODO: Make page length a configuration parameter that is also
	//       used on the approval page.
	dbchs, err := ctx.db.GetApprovedChannels((page - 1) * 15, 15 + 1)
	if err != nil {
		panic(err)
	}

	if len(dbchs) > 15 {
		dbchs = dbchs[:15]
		ctx.index.MoreNext = true
	}

	ctx.index.MorePrev = page > 1
	ctx.index.PageNext = page + 1
	ctx.index.PagePrev = page - 1
	ctx.index.Channels = dbchs

	ctx.index.initialized = true

	return &ctx.index
}

func (ctx *serveContext) serveIndex(w http.ResponseWriter, req *http.Request) {
	c, err := dbOpen()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	page := 1
	pagestr := req.URL.Query().Get("page")
	if pagestr != "" {
		fmt.Sscanf(pagestr, "%d", &page)
		if page < 1 {
			page = 1
		}
	}

	if page == 1 {
		ctx.setPageTitle(conf.WebsiteTitle)
	} else {
		ctx.setPageTitle(fmt.Sprintf("Side %d", page))
	}

	err = ctx.executeTemplate(w, "index", ctx)
	if err != nil {
		panic(err)
	}
}

