package main

import (
	"net/http"
	"html/template"
	"fmt"
)

func errorServe(w http.ResponseWriter, req *http.Request, msg string) {
	data := struct{
		serveCommon
		PageTitle string
		Admin bool
		Message string
	}{
		serveCommon: serveCommonData(req),
		PageTitle: "Feilmelding",
		Message: msg,
	}

	tpath := conf.AssetsPath + "/templates.html"
	t, err := template.ParseFiles(tpath)
	if err != nil {
		panic(fmt.Errorf("Failed to parse template file '%s': %s.\n", tpath, err.Error()))
	}

	err = t.ExecuteTemplate(w, "error", &data)
	if err != nil {
		panic(fmt.Errorf("Failed to execute template file '%s': %s.\n", tpath, err.Error()))
	}
}
