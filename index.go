package main

import (
	"net/http"
	"html/template"
	"log"
)

func indexServe(conf *config, w http.ResponseWriter, req *http.Request) {
	tpath := conf.AssetsPath + "/index.html"
	t, err := template.ParseFiles(tpath)
	if err != nil {
		log.Printf("Failed to parse template file '%s': %s.\n", tpath, err.Error())
	}
	t.Execute(w, nil)
}
