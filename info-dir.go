package main

import (
	"bytes"
	"github.com/ivartj/norske-irc-kanaler.com/bbgo"
	"html/template"
	"net/http"
	"os"
	"path"
)

func infoDir(page *page, req *http.Request) {
	topic := path.Base(req.URL.Path)
	filepath := page.main.conf.AssetsPath() + "/info/" + topic + ".txt"
	file, err := os.Open(filepath)
	if err != nil {
		page.SetField("content", "Fant ikke noen slik side.")
		page.WriteHeader(404)
	} else {
		defer file.Close()
		buf := bytes.NewBuffer([]byte{})
		err = bbgo.Process(file, buf)
		if err != nil {
			page.Fatalf("Failed to convert file from BBCode to HTML: %s", err.Error())
		}
		page.SetField("content", template.HTML(buf.String()))
	}

	page.ExecuteTemplate("info")
}
