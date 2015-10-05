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

func (ctx *serveContext) serveHelp(w http.ResponseWriter, req *http.Request) {
	data := struct{
		*serveContext
		Content template.HTML
	}{
		serveContext: ctx,
		Content: helpGetContent(),
	}

	ctx.setPageTitle("Hjelp med IRC-chat")

	err := ctx.executeTemplate(w, "help", &data)
	if err != nil {
		panic(err)
	}
}
