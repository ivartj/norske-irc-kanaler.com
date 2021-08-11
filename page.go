package main

import (
	"bytes"
	"fmt"
	"github.com/ivartj/norske-irc-kanaler.com/bbgo"
	"github.com/ivartj/norske-irc-kanaler.com/web"
	"html/template"
	"net/http"
	"os"
	"strings"
)

type page struct {
	web.Page
	main *mainContext
}

func pageHandler(ctx *mainContext, pageFn func(*page, *http.Request)) func(web.Page, *http.Request) {
	return func(webPage web.Page, req *http.Request) {
		pg := &page{
			main: ctx,
			Page: webPage,
		}

		if req.Referer() != "" {
			pg.SetField("referer", req.Referer())
		} else {
			pg.SetField("referer", "/")
		}

		pg.SetField("site-stylesheet-modtime", "")
		info, err := os.Stat(ctx.conf.AssetsPath() + "/static/styles.css")
		if err == nil {
			pg.SetField("site-stylesheet-modtime", info.ModTime().Unix())
		}

		ctx.auth.InitPage(pg, req)

		pageFn(pg, req)
	}
}

func (page *page) AddMessage(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	bb := bytes.NewBuffer([]byte{})
	err := bbgo.Process(strings.NewReader(msg), bb)
	if err != nil {
		page.Fatalf("Failed to convert '%s' from BBCode to HTML: %s", err.Error())
	}
	imsgs, err := page.GetField("page-messages")
	if err != nil {
		page.Fatalf("page-messages field not set")
	}
	msgs, ok := imsgs.([]template.HTML)
	if !ok {
		fmt.Printf("%T", imsgs)
		page.Fatalf("Failed to convert field data for 'page-messages' into []template.HTML.")
	}
	page.SetField("page-messages", append(msgs, template.HTML(bb.String())))
}
