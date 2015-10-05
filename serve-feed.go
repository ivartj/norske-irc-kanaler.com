package main

import (
	"net/http"
	"fmt"
	"html"
	"time"
)

func (ctx *serveContext) serveFeed(w http.ResponseWriter, req *http.Request) {
	c, err := dbOpen()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	chs, err := c.GetApprovedChannels(0, 15)
	if err != nil {
		panic(err)
	}

	feedServeCommon(w, req, chs, "Norske IRC-kanaler")
}

func (ctx *serveContext) serveFeedUnapproved(w http.ResponseWriter, req *http.Request) {
	c, err := dbOpen()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	chs, err := c.GetUnapprovedChannels(0, 15)
	if err != nil {
		panic(err)
	}

	feedServeCommon(w, req, chs, "Ikke-godkjente norske IRC-kanaler")
}

func feedServeCommon(w http.ResponseWriter, req *http.Request, chs []channel, title string) {
	fmt.Fprintln(w, `<?xml version="1.0"?>
<rss version="2.0">
	<channel>
		<title>` + html.EscapeString(title) +  `</title>
		<link>http://` + html.EscapeString(req.Host) + `</link>
		<language>no</language>
	`)

	for _, v := range chs {
		nameAndServer :=  fmt.Sprintf("%s@%s", v.Name(), v.Network())
		fmt.Fprintln(w, `
		<item>
			<title>` + html.EscapeString(nameAndServer) + `</title>
			<link>` + html.EscapeString(v.Weblink()) + `</link>
			<description>` + html.EscapeString(v.Description()) + `</description>
			<pubdate>` + html.EscapeString(v.ApproveTime().Format(time.RFC1123Z)) + `</pubdate>
			<guid>` + html.EscapeString(nameAndServer) + `</guid>
		</item>
		`)
	}

	fmt.Fprintln(w, `
	</channel>
</rss>
	`)
}
