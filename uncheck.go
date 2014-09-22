package main

import (
	"net/http"
	"net/url"
	"fmt"
	"html/template"
)

func uncheckServe(w http.ResponseWriter, req *http.Request) {
	if loginAuth(req) == false {
		http.Redirect(w, req, "/login?redirect=" + url.QueryEscape(req.URL.Path + "?" + req.URL.RawQuery), 307)
		return
	}

	data := struct{
		PageTitle string
		Name string
		Server string
		Message string
		Redirect string
	}{
		PageTitle: "Sletting av kanal",
		Name: req.URL.Query().Get("name"),
		Server: req.URL.Query().Get("server"),
	}

	c := dbOpen()
	defer c.Close()

	dbUncheck(c, data.Name, data.Server)
	data.Message = fmt.Sprintf("%s@%s skal bli omskjekket.", data.Name, data.Server)
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
