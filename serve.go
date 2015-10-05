package main

import (
	"fmt"
	"os"
	"io"
	"log"
	"runtime/debug"
	"net"
	"net/http"
	"net/http/fcgi"
	"html/template"
)

type serveContext struct{
	req *http.Request
	pageTitle string
}

func newServeContext(req *http.Request) *serveContext {
	return &serveContext{
		req: req,
		pageTitle: conf.WebsiteTitle,
	}
}

// Initiated in serve()
var serveTemplate *template.Template = nil

func (ctx *serveContext) SiteTitle() string {
	return conf.WebsiteTitle
}

func (ctx *serveContext) SiteDescription() string {
	return conf.WebsiteDescription
}

func (ctx *serveContext) PageTitle() string {
	return ctx.pageTitle
}

func (ctx *serveContext) setPageTitle(title string) {
	ctx.pageTitle = title
}

func (ctx *serveContext) Admin() bool {
	return loginAuth(ctx.req)
}

func (ctx *serveContext) executeTemplate(w io.Writer, name string, data interface{}) error {
	var t *template.Template
	var err error

	if conf.ReloadTemplate {
		tpath := conf.AssetsPath + "/templates.html"
		t, err = template.ParseFiles(tpath)
		if err != nil {
			return err
		}
	} else {
		t = serveTemplate
	}

	err = t.ExecuteTemplate(w, name, &data)
	if err != nil {
		return err
	}

	return nil
}

func serveRecovery(w http.ResponseWriter, req *http.Request) {
	defer func() {
		x := recover()
		if x != nil {
			msg := fmt.Sprintf("%s: %s\n", x, debug.Stack())
			log.Printf("%s", msg)
			ctx := newServeContext(req)
			ctx.serveError(w, req, msg)
		}
	}()
	http.DefaultServeMux.ServeHTTP(w, req)
}

func serveExact(w http.ResponseWriter, req *http.Request) {
	ctx := newServeContext(req)
	switch req.URL.Path {
	case "/":
		ctx.serveIndex(w, req)
	case "/submit":
		ctx.serveSubmit(w, req)
	case "/feed":
		ctx.serveFeed(w, req)
	case "/feed-unapproved":
		ctx.serveFeedUnapproved(w, req)
	case "/login":
		ctx.serveLogin(w, req)
	case "/logout":
		ctx.serveLogout(w, req)
	case "/edit":
		ctx.serveEdit(w, req)
	case "/approve":
		ctx.serveApprove(w, req)
	case "/delete":
		ctx.serveDelete(w, req)
	case "/help":
		ctx.serveHelp(w, req)
	case "/favicon.ico":
		http.ServeFile(w, req, conf.AssetsPath + "/favicon.ico")
	default:
		http.NotFound(w, req)
	}
}

func serve() {
	var err error
	serveTemplate, err = template.ParseFiles(conf.AssetsPath + "/templates.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %s.\n", err.Error())
	}

	http.DefaultServeMux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(conf.AssetsPath + "/static"))))
	http.DefaultServeMux.HandleFunc("/", serveExact)

	switch conf.ServeMethod {
	case "http":
		log.Fatal(http.ListenAndServe(":" + fmt.Sprintf("%d", conf.HttpPort), http.HandlerFunc(serveRecovery)))
	case "fastcgi":

		// TODO check that it is a socket
		os.Remove(conf.FastcgiPath)

		l, err := net.Listen("unix", conf.FastcgiPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to listen on Fastcgi path '%s': %s.\n", conf.FastcgiPath, err.Error())
			os.Exit(1)
		}
		log.Fatal(fcgi.Serve(l, http.HandlerFunc(serveRecovery)))
	}
}

