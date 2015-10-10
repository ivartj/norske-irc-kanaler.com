package main

import (
	"net/http"
	"html/template"
	"fmt"
	"os"
	"path"
	"io/ioutil"
	"github.com/frustra/bbcode"
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

type serveInfoContext struct {
	initialized bool
	Content template.HTML
}

func (ctx *serveContext) Info() *serveInfoContext {
	if ctx.info.initialized {
		return &ctx.info
	}

	topic := path.Base(ctx.req.URL.Path)
	if topic == "/" {
		topic = "help"
	}

	filepath := conf.AssetsPath + "/info/" + topic + ".txt"
	file, err := os.Open(filepath)
	if err == os.ErrNotExist {
		ctx.setMessage("No such topic.")
		ctx.w.WriteHeader(404)
	} else if err != nil {
		panic(err)
	} else {
		defer file.Close()
		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}

		bb := bbcode.NewCompiler(true, true)
		ctx.info.Content = template.HTML(bb.Compile(string(bytes)))
	}

	ctx.info.initialized = true
	return &ctx.info
}

func (ctx *serveContext) serveInfo(w http.ResponseWriter, req *http.Request) {
	// Needs to be called immediately in case of 404
	ctx.Info()
	ctx.setPageTitle("IRC-info")
	err := ctx.executeTemplate(w, "info", ctx)
	if err != nil {
		panic(err)
	}
}

