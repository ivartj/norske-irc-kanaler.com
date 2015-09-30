package main

import (
	"net/http"
	"net/url"
	"fmt"
	"html/template"
)

func deleteServe(w http.ResponseWriter, req *http.Request) {
	if loginAuth(req) == false {
		http.Redirect(w, req, "/login?redirect=" + url.QueryEscape(req.URL.Path + "?" + req.URL.RawQuery), 307)
		return
	}

	data := struct{
		serveCommon
		PageTitle string
		Name string
		Server string
		Message string
		Redirect string
	}{
		serveCommon: serveCommonData(req),
		PageTitle: "Sletting av kanal",
		Name: req.URL.Query().Get("name"),
		Server: req.URL.Query().Get("server"),
	}

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

	tpath := conf.AssetsPath + "/templates.html"
	t, err := template.ParseFiles(tpath)
	if err != nil {
		panic(fmt.Errorf("Failed to parse template file '%s': %s.\n", tpath, err.Error()))
	}

	err = t.ExecuteTemplate(w, "delete", &data)
	if err != nil {
		panic(fmt.Errorf("Failed to execute template file '%s': %s.\n", tpath, err.Error()))
	}
}
