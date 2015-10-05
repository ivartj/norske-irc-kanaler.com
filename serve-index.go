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

func (ctx *serveContext) serveIndex(w http.ResponseWriter, req *http.Request) {
	c, err := dbOpen()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	data := struct{
		*serveContext
		Channels []indexChannel
		MoreNext bool
		MorePrev bool
		PageNext int
		PagePrev int
	}{
		serveContext: ctx,
	}

	ctx.setPageTitle(conf.WebsiteTitle)

	pagestr := req.URL.Query().Get("page")
	page := 1
	fmt.Sscanf(pagestr, "%d", &page)
	if page < 1 {
		page = 1
	}

	if page != 1 {
		ctx.setPageTitle(fmt.Sprintf("Side %d", page))
	}

	// TODO: Make page length a configuration parameter that is also
	//       used on the approval page.
	dbchs, err := c.GetApprovedChannels((page - 1) * 15, 15 + 1)
	if err != nil {
		panic(err)
	}

	if len(dbchs) > 15 {
		dbchs = dbchs[:15]
		data.MoreNext = true
	}

	data.MorePrev = page > 1
	data.PageNext = page + 1
	data.PagePrev = page - 1

	chs := make([]indexChannel, len(dbchs))
	for i, v := range dbchs {
		status, ok := channelStatusString(v)
		chs[i] = indexChannel{
			Name: v.Name(),
			Server: v.Network(),
			WebLink: v.Weblink(),
			Description: v.Description(),
			Status: status,
			Error: !ok,
		}
	}
	data.Channels = chs

	err = ctx.executeTemplate(w, "index", &data)
	if err != nil {
		panic(err)
	}
}

