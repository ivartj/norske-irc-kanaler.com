package main

import (
	"net/http"
	"net/url"
	"fmt"
)

func (ctx *serveContext) serveDelete(w http.ResponseWriter, req *http.Request) {
	if loginAuth(req) == false {
		http.Redirect(w, req, "/login?redirect=" + url.QueryEscape(req.URL.Path + "?" + req.URL.RawQuery), 307)
		return
	}

	data := struct{
		*serveContext
		Name string
		Server string
		Message string
		Redirect string
	}{
		serveContext: ctx,
		Name: req.URL.Query().Get("name"),
		Server: req.URL.Query().Get("server"),
	}

	ctx.setPageTitle("Sletting av kanal")

	c, err := dbOpen()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	err = c.DeleteChannel(data.Name, data.Server)
	if err != nil {
		panic(err)
	}
	data.Message = fmt.Sprintf("%s@%s har blitt slettet.", data.Name, data.Server)
	data.Redirect = req.Referer()

	err = ctx.executeTemplate(w, "delete", &data)
	if err != nil {
		panic(err)
	}
}

