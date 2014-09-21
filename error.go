package main

import (
	"net/http"
	"html/template"
	"log"
)

func errorServe(w http.ResponseWriter, req *http.Request, msg string) {
	data := struct{
		PageTitle string
		Admin bool
		Message string
	}{
		PageTitle: "Feilmelding",
		Admin: loginAuth(req),
		Message: msg,
	}

	tpath := conf.AssetsPath + "/templates.html"
	t, err := template.ParseFiles(tpath)
	if err != nil {
		log.Panicf("Failed to parse template file '%s': %s.\n", tpath, err.Error())
	}

	err = t.ExecuteTemplate(w, "error", &data)
	if err != nil {
		log.Panicf("Failed to execute template file '%s': %s.\n", tpath, err.Error())
	}
}
