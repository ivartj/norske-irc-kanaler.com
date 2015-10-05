package main

import (
	"net/http"
	"fmt"
	"net/url"
)

func (ctx *serveContext) serveEdit(w http.ResponseWriter, req *http.Request) {
	if loginAuth(req) == false {
		http.Redirect(w, req, "/login?redirect=" + url.QueryEscape(req.URL.Path + "?" + req.URL.RawQuery), 307)
		return
	}

	data := struct{
		*serveContext
		OriginalName string
		OriginalServer string
		Name string
		Server string
		WebLink string
		Description string
		Message string
	}{
		serveContext: ctx,
	}

	ctx.setPageTitle("Rediger kanal")

	c, err := dbOpen()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	switch req.Method {
	case "GET":
		q := req.URL.Query()
		data.OriginalName = q.Get("name")
		data.Name = q.Get("name")
		data.OriginalServer = q.Get("server")
		data.Server = q.Get("server")
		ch, err := c.GetChannel(data.OriginalName, data.OriginalServer)
		if err != nil {
			panic(fmt.Errorf("Failed to retrieve channel %s@%s from database: %s", data.OriginalName, data.OriginalServer, err.Error()))
		}
		data.WebLink = ch.Weblink()
		data.Description = ch.Description()
	case "POST":
		data.OriginalName = req.FormValue("name")
		data.Name = req.FormValue("name")
		data.OriginalServer = req.FormValue("server")
		data.Server = req.FormValue("server")
		data.WebLink = req.FormValue("weblink")
		data.Description = req.FormValue("description")
		err = c.EditChannel(
			req.FormValue("original-name"),
			req.FormValue("original-server"),
			data.Name,
			data.Server,
			data.WebLink,
			data.Description)
		if err != nil {
			panic(err)
		}
		data.Message = "Endring vellykket."
	}

	err = ctx.executeTemplate(w, "edit", &data)
	if err != nil {
		panic(err)
	}
}
