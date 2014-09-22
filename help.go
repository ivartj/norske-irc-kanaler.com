package main

import (
	"net/http"
	"html/template"
	"fmt"
	"os"
	"io/ioutil"
)

func helpGetContent() template.HTML {
	hpath := conf.AssetsPath + "/help.html"
	f, err := os.Open(hpath)
	if err != nil {
		panic(fmt.Errorf("Failed to open file '%s': %s", hpath, err.Error()))
	}
	defer f.Close()
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		panic(fmt.Errorf("Error on reading file '%s': %s", hpath, err.Error()))
	}
	return template.HTML(string(bytes))
}

func helpServe(w http.ResponseWriter, req *http.Request) {
	data := struct{
		PageTitle string
		Content template.HTML
	}{
		PageTitle: "Hjelp med IRC-chat",
		Content: helpGetContent(),
	}

	tpath := conf.AssetsPath + "/templates.html"
	t, err := template.ParseFiles(tpath)
	if err != nil {
		panic(fmt.Errorf("Failed to parse template file '%s': %s", tpath, err.Error()))
	}

	err = t.ExecuteTemplate(w, "help", &data)
	if err != nil {
		panic(fmt.Errorf("Failed to execute template file '%s': %s", tpath, err.Error()))
	}
}
