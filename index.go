package main

import (
	"net/http"
	"html/template"
	"log"
	"fmt"
	sql "code.google.com/p/go-sqlite/go1/sqlite3"
)

type indexChannel struct{
	Name string
	Server string
	WebLink string
	Description string
	Status string
	Approved bool
}

func indexGetChannels(c *sql.Conn) ([]indexChannel, bool) {
	more := false

	dbchs, numtotal := dbGetApprovedChannels(c, 0, 15)
	if numtotal > 15 {
		more = true
	}

	chs := make([]indexChannel, len(dbchs))
	for i, v := range dbchs {
		status := ""
		if v.errmsg == "" {
			if v.numusers == 1 {
				status = "1 bruker innlogget"
			} else {
				status = fmt.Sprintf("%d brukere innlogget", v.numusers)
			}
			status += " " + timeAgo(v.lastcheck)
		} else {
			status = v.errmsg + " " + timeAgo(v.lastcheck)
		}
		chs[i] = indexChannel{
			Name: v.name,
			Server: v.server,
			WebLink: v.weblink,
			Description: v.description,
			Approved: v.approved,
			Status: status,
		}
	}

	return chs, more
}

func indexServe(w http.ResponseWriter, req *http.Request) {
	c := dbOpen()
	defer c.Close()

	channels, more := indexGetChannels(c)
	data := struct{
		PageTitle string
		Channels []indexChannel
		More bool
	}{
		PageTitle: "IRC-Chat Norge",
		Channels: channels,
		More: more,
	}

	tpath := conf.AssetsPath + "/templates.html"
	t, err := template.ParseFiles(tpath)
	if err != nil {
		log.Panicf("Failed to parse template file '%s': %s.\n", tpath, err.Error())
	}
	err = t.ExecuteTemplate(w, "index", &data)
	if err != nil {
		log.Panicf("Failed to execute template file '%s': %s.\n", tpath, err.Error())
	}
}
