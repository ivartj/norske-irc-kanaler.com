package main

import (
	"net/http"
	"strconv"
	"github.com/ivartj/norske-irc-kanaler.com/web"
)

const approvePageSize = 15

type serveApproveContext struct {
	initialized bool
	Channels []channel
	ApproveName string
	ApproveNetwork string
	MoreNext bool
	MorePrev bool
	PageNext int
	PagePrev int
}

func approvePage(page web.Page, req *http.Request) {

	page.SetField("page-title", "Kanalgodkjenning")

	approveName := req.URL.Query().Get("name")
	approveNetwork := req.URL.Query().Get("network")
	if approveName != "" && approveNetwork != "" {
		err := dbApproveChannel(page, approveName, approveNetwork)
		if err != nil {
			page.Fatalf("Failed to approve channel: %s", err.Error())
		}
		utilAddMessage(page, "Kanalen er godkjent!")
	}
	pgstr := req.URL.Query().Get("page")
	pg := 1
	pg, err := strconv.Atoi(pgstr)
	if err != nil || pg < 1 {
		pg = 1
	}

	rows, err := page.Query(
		"select * from channel_unapproved limit ? offset ?;",
		approvePageSize + 1, (pg - 1) * approvePageSize) 
	if err != nil {
		page.Fatalf("Failed to query unapproved channels: %s", err.Error())
	}

	chs, err := dbScanChannels(rows)
	if err != nil {
		page.Fatalf("Failed to scan queried channels: %s", err.Error())
	}

	moreNext := false
	if len(chs) > approvePageSize {
		chs = chs[:approvePageSize]
		moreNext = true
	}

	page.SetFieldMap(map[string]interface{}{
		"more-next" : moreNext,
		"more-prev" : pg > 1,
		"page-next" : pg + 1,
		"page-prev" : pg - 1,
		"channels" : chs,
	})

	page.ExecuteTemplate("approve")
}

