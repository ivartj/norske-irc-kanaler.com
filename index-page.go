package main

import (
	"strconv"
	"net/http"
	"fmt"
)

const (
	indexPageSize = 15
)

func indexPage(page *page, req *http.Request) {

	var err error

	pg := 1
	pgstr := req.URL.Query().Get("page")
	if pgstr != "" {
		pg, err = strconv.Atoi(pgstr)
		if err != nil || pg < 1 {
			pg = 1
		}
	}

	if pg != 1 {
		page.SetField("page-title", fmt.Sprintf("Side %d", pg))
	}

	rows, err := page.Query(
		"select * from channel_indexed limit ? offset ?;",
		indexPageSize + 1, (pg - 1) * indexPageSize) 
	if err != nil {
		page.Fatalf("Failed to query index channels: %s", err.Error())
	}

	chs, err := dbScanChannels(rows)
	if err != nil {
		page.Fatalf("Failed to scan queried channels: %s", err.Error())
	}

	moreNext := false
	if len(chs) > indexPageSize {
		moreNext = true
		chs = chs[:indexPageSize]
	}

	page.SetFieldMap(map[string]interface{}{
		"more-prev" : pg > 1,
		"more-next" : moreNext,
		"page-next" : pg + 1,
		"page-prev" : pg - 1,
		"channels" : chs,
	})

	page.ExecuteTemplate("index")
}

