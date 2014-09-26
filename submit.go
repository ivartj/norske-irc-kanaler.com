package main

import (
	"net/http"
	"html/template"
	"fmt"
	sql "code.google.com/p/go-sqlite/go1/sqlite3"
)

func submitChannel(c *sql.Conn, name, server, weblink, description string) string {

	name, server = chanCanonical(name, server)

	if weblink == "" {
		weblink = chanSuggestWebLink(name, server)
	}

	err := chanValidate(name, server)
	if err != nil {
		return err.Error()
	}

	ch := dbGetChannel(c, name, server)
	if ch != nil {
		return "Takk. Forslaget har allerede blitt sendt inn."
	}

	dbAddChannel(c, name, server, weblink, description, 0)

	if conf.Approval {
		return "Takk for forslag! Forslaget vil publiseres etter godkjenning av administrator."
	} else {
		return "Takk for forslag! Forslaget er publisert."
	}
}

func submitServe(w http.ResponseWriter, req *http.Request) {
	data := struct{
		serveCommon
		PageTitle string
		Name string
		Server string
		WebLink string
		Description string
		Message string
	}{
		serveCommon: serveCommonData(req),
		PageTitle: "Legg til IRC-chatterom",
	}

	if req.Method == "POST" {
		data.Name = req.FormValue("name")
		data.Server = req.FormValue("server")
		data.WebLink = req.FormValue("weblink")
		data.Description = req.FormValue("description")
		c := dbOpen()
		defer c.Close()

		data.Message = submitChannel(c, data.Name, data.Server, data.WebLink, data.Description)
	}

	tpath := conf.AssetsPath + "/templates.html"
	t, err := template.ParseFiles(tpath)
	if err != nil {
		panic(fmt.Errorf("Failed to parse template file '%s': %s.\n", tpath, err.Error()))
	}
	err = t.ExecuteTemplate(w, "submit", &data)
	if err != nil {
		panic(fmt.Errorf("Failed to execute template file '%s': %s.\n", tpath, err.Error()))
	}
}
