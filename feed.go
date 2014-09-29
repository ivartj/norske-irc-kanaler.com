package main

import (
	"net/http"
	"fmt"
	"html"
	"time"
)

func feedServe(w http.ResponseWriter, req *http.Request) {
	c := dbOpen()
	defer c.Close()

	chs, _ := dbGetApprovedChannels(c, 0, 15)
	feedServeCommon(w, req, chs, "Norske IRC-kanaler")
}

func feedUnapprovedServe(w http.ResponseWriter, req *http.Request) {
	c := dbOpen()
	defer c.Close()

	chs, _ := dbGetUnapprovedChannels(c, 0, 15)
	feedServeCommon(w, req, chs, "Ikke-godkjente norske IRC-kanaler")
}

func feedServeCommon(w http.ResponseWriter, req *http.Request, chs []dbChannel, title string) {
	fmt.Fprintln(w, `<?xml version="1.0"?>
<rss version="2.0">
	<channel>
		<title>` + html.EscapeString(title) +  `</title>
		<link>http://` + html.EscapeString(req.Host) + `</link>
		<language>no</language>
	`)

	for _, v := range chs {
		nameAndServer :=  fmt.Sprintf("%s@%s", v.name, v.server)
		fmt.Fprintln(w, `
		<item>
			<title>` + html.EscapeString(nameAndServer) + `</title>
			<link>` + html.EscapeString(v.weblink) + `</link>
			<description>` + html.EscapeString(v.description) + `</description>
			<pubdate>` + html.EscapeString(v.approvedate.Format(time.RFC1123Z)) + `</pubdate>
			<guid>` + html.EscapeString(nameAndServer) + `</guid>
		</item>
		`)
	}

	fmt.Fprintln(w, `
	</channel>
</rss>
	`)
}
