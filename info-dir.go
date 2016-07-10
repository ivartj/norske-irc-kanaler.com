package main

import (
	"net/http"
	"html/template"
	"os"
	"path"
	"bytes"
	"github.com/ivartj/norske-irc-kanaler.com/bbgo"
	"github.com/ivartj/norske-irc-kanaler.com/web"
)

func infoDir(page web.Page, req *http.Request) {
	topic := path.Base(req.URL.Path)
	filepath := conf.AssetsPath + "/info/" + topic + ".txt"
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

