package main

import (
	"net/http"
	"html/template"
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

func indexServe(w http.ResponseWriter, req *http.Request) {
	c := dbOpen()
	defer c.Close()

	data := struct{
		PageTitle string
		Channels []indexChannel
		MoreNext bool
		MorePrev bool
		PageNext int
		PagePrev int
		Admin bool
	}{
		PageTitle: "IRC-Chat Norge",
		Admin: loginAuth(req),
	}

	pagestr := req.URL.Query().Get("page")
	page := 1
	fmt.Sscanf(pagestr, "%d", &page)
	if page < 1 {
		page = 1
	}

	if page != 1 {
		data.PageTitle = fmt.Sprintf("Side %d", page)
	}

	dbchs, total := dbGetApprovedChannels(c, (page - 1) * 15, 15)
	data.MoreNext = total > page * 15
	data.MorePrev = page > 1
	data.PageNext = page + 1
	data.PagePrev = page - 1

	chs := make([]indexChannel, len(dbchs))
	for i, v := range dbchs {
		status, ok := chanStatus(&v)
		chs[i] = indexChannel{
			Name: v.name,
			Server: v.server,
			WebLink: v.weblink,
			Description: v.description,
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
	err = t.ExecuteTemplate(w, "index", &data)
	if err != nil {
		panic(fmt.Errorf("Failed to execute template file '%s': %s.\n", tpath, err.Error()))
	}
}
