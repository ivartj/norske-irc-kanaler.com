package main

import (
	"fmt"
	"os"
	"io"
	"strings"
	"log"
	"runtime/debug"
	"net"
	"net/http"
	"net/http/fcgi"
	"html/template"
	"github.com/ivartj/norske-irc-kanaler.com/bbgo"
	"bytes"
)

type serveContext struct{
	index serveIndexContext
	submit serveSubmitContext
	approve serveApproveContext
	exclude serveExcludeContext
	adminpanel serveAdminPanelContext
	info serveInfoContext
	edit serveEditContext
	req *http.Request
	w http.ResponseWriter
	pageTitle string
	message template.HTML
	db *dbConn
}

func newServeContext(w http.ResponseWriter, req *http.Request) (*serveContext, error) {
	db, err := dbOpen()
	if err != nil {
		return nil, err
	}
	return &serveContext{
		w: w,
		req: req,
		pageTitle: conf.WebsiteTitle,
		db: db,
	}, nil
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

func (ctx *serveContext) SessionID() string {
	return loginSessionID
}

func (ctx *serveContext) Message() template.HTML {
	return ctx.message
}

func (ctx *serveContext) setMessage(msg string) {
	output := bytes.NewBuffer([]byte{})
	input := strings.NewReader(msg)
	err := bbgo.Process(input, output)
	if err != nil {
		panic(err)
	}
	ctx.message = template.HTML(output.String())
}

func (ctx *serveContext) executeTemplate(w io.Writer, name string, data interface{}) error {
	var t *template.Template
	var err error

	fnmap := template.FuncMap{
		"context": func() *serveContext { return ctx },
	}

	if conf.ReloadTemplate {
		tpath := conf.AssetsPath + "/templates.html"
		t, err = template.New(name).Funcs(fnmap).ParseFiles(tpath)
		if err != nil {
			return err
		}
	} else {
		t = serveTemplate
		t.Funcs(fnmap)
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

			// TODO: Do not depend serveContext
			ctx, _ := newServeContext(w, req)
			ctx.serveError(w, req, msg)
		}
	}()
	http.DefaultServeMux.ServeHTTP(w, req)
}

func serveExact(w http.ResponseWriter, req *http.Request) {
	ctx, err := newServeContext(w, req)
	if err != nil {
		panic(err)
	}

	if strings.HasPrefix(req.URL.Path, "/info") {
		ctx.serveInfo(w, req)
		return
	}

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
	case "/admin":
		ctx.serveAdmin(w, req)
	case "/approve":
		ctx.serveApprove(w, req)
	case "/exclude":
		ctx.serveExclude(w, req)
	case "/delete":
		ctx.serveDelete(w, req)
	case "/favicon.ico":
		http.ServeFile(w, req, conf.AssetsPath + "/favicon.ico")
	default:
		http.NotFound(w, req)
	}
}

func serve() {
	var err error

	fnmap := template.FuncMap{
		"context": func() *serveContext { return nil },
	}

	serveTemplate, err = template.New("main").Funcs(fnmap).ParseFiles(conf.AssetsPath + "/templates.html")
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

