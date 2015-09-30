package main

import (
	"net/http"
	"html/template"
	"fmt"
	"net/url"
)

type approveChannel struct {
	Name string
	Server string
	WebLink string
	Description string
	Status string
	Error bool
}

func approveServe(w http.ResponseWriter, req *http.Request) {

	if loginAuth(req) == false {
		http.Redirect(w, req, "/login?redirect=" + url.QueryEscape(req.URL.Path + "?" + req.URL.RawQuery), 307)
		return
	}

	data := struct{
		serveCommon
		PageTitle string
		Channels []approveChannel
		Admin bool
		ApproveName string
		ApproveServer string
		Message string
		MoreNext bool
		MorePrev bool
		PageNext int
		PagePrev int
	} {
		serveCommon: serveCommonData(req),
		PageTitle: "Kanalgodkjenning",
		Admin: loginAuth(req),
	}

	c, err := dbOpen()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	data.ApproveName = req.URL.Query().Get("name")
	data.ApproveServer = req.URL.Query().Get("server")
	if data.ApproveName != "" && data.ApproveServer != "" {
		err = c.ApproveChannel(data.ApproveName, data.ApproveServer)
		if err != nil {
			panic(err)
		}
		data.Message = "Kanalen er godkjent!"
	}

	pagestr := req.URL.Query().Get("page")
	page := 1
	fmt.Sscanf(pagestr, "%d", &page)
	if page < 1 {
		page = 1
	}

	dbchs, err := c.GetUnapprovedChannels((page - 1) * 15, 15 + 1)
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

	chs := make([]approveChannel, len(dbchs))
	for i, v := range dbchs {
		status, ok := channelStatusString(v)
		chs[i] = approveChannel{
			Name: v.Name(),
			Server: v.Network(),
			WebLink: v.Weblink(),
			Description: v.Description(),
			Status: status,
			Error: !ok,
		}
	}

	data.Channels = chs

	tpath := conf.AssetsPath + "/templates.html"
	t, err := template.ParseFiles(tpath)
	if err != nil {
		panic(fmt.Errorf("Failed to parse template file '%s': %s.\n", tpath, err.Error()))
	}

	err = t.ExecuteTemplate(w, "approve", &data)
	if err != nil {
		panic(fmt.Errorf("Failed to execute template file '%s': %s.\n", tpath, err.Error()))
	}
}

