package main

import (
	"net/http"
	"html/template"
	"fmt"
)

func submitChannel(c *dbConn, name, server, weblink, description string) string {

	name, server = channelAddressCanonical(name, server)

	if weblink == "" {
		weblink = channelSuggestWebLink(name, server)
	}

	err := channelAddressValidate(name, server)
	if err != nil {
		return err.Error()
	}

	isExcluded, excludeReason, err := c.IsChannelExcluded(name, server)
	if err != nil {
		panic(err)
	}

	if isExcluded {
		return fmt.Sprintf("Kanalen blir ikke opplistet av følgende grunn: %s.\n", excludeReason)
	}

	ch, _ := c.GetChannel(name, server)
	if ch != nil {
		return "Takk. Bidraget har allerede blitt sendt inn."
	}

	err = c.AddChannel(name, server, weblink, description, !conf.Approval)
	if err != nil {
		panic(err)
	}

	if conf.Approval {
		return "Takk for forslaget! Forslaget vil publiseres etter godkjenning av administrator."
	} else {
		return "Takk for bidraget! Forslaget er publisert."
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
		c, err := dbOpen()
		if err != nil {
			panic(err)
		}
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

