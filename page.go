package main

import (
	"github.com/ivartj/norske-irc-kanaler.com/web"
	"github.com/ivartj/norske-irc-kanaler.com/bbgo"
	"fmt"
	"bytes"
	"strings"
	"html/template"
)

type page struct {
	web.Page
	main *mainContext
}

func pageNew(ctx *mainContext, webPage web.Page) *page {
	return &page{
		main: ctx,
		Page: webPage,
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

