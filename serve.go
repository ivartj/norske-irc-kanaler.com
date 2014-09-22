package main

import (
	"fmt"
	"os"
	"log"
	"runtime/debug"
	"net"
	"net/http"
	"net/http/fcgi"
)

func serveRecovery(w http.ResponseWriter, req *http.Request) {
	defer func() {
		x := recover()
		if x != nil {
			msg := fmt.Sprintf("%s: %s\n", x, debug.Stack())
			log.Printf("%s", msg)
			errorServe(w, req, msg)
		}
	}()
	http.DefaultServeMux.ServeHTTP(w, req)
}

func serveExact(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/":
		indexServe(w, req)
	case "/submit":
		submitServe(w, req)
	case "/login":
		loginServe(w, req)
	case "/logout":
		logoutServe(w, req)
	case "/edit":
		editServe(w, req)
	case "/approve":
		approveServe(w, req)
	case "/delete":
		deleteServe(w, req)
	case "/uncheck":
		uncheckServe(w, req)
	case "/help":
		helpServe(w, req)
	default:
		http.NotFound(w, req)
	}
}

func serve() {
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
