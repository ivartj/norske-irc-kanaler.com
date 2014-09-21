package main

import (
	"net/http"
	"html/template"
	"log"
)

func submitServe(w http.ResponseWriter, req *http.Request) {
	data := struct{
		PageTitle string
		Name string
		Server string
		WebLink string
		Description string
		Message string
	}{
		PageTitle: "Foreslå IRC-chatterom",
	}

	if req.Method == "POST" {
		data.Name = req.FormValue("name")
		data.Server = req.FormValue("server")
		data.WebLink = req.FormValue("weblink")
		data.Description = req.FormValue("description")
		c := dbOpen()
		defer c.Close()

		ch := dbGetChannel(c, data.Name, data.Server)
		if ch == nil {
			dbAddChannel(c, data.Name, data.Server, data.WebLink, data.Description, 0)
			data.Message = "Takk for forslag! Forslaget vil publiseres etter godkjenning av administrator."
		} else {
			data.Message = "Takk. Forslaget har allerede blitt sendt inn."
		}
	}

	tpath := conf.AssetsPath + "/templates.html"
	t, err := template.ParseFiles(tpath)
	if err != nil {
		log.Panicf("Failed to parse template file '%s': %s.\n", tpath, err.Error())
	}
	err = t.ExecuteTemplate(w, "submit", &data)
	if err != nil {
		log.Panicf("Failed to execute template file '%s': %s.\n", tpath, err.Error())
	}
}
