package main

import (
	"fmt"
	"os"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
)

func serveExact(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/":
		indexServe(w, req)
	case "/submit":
		submitServe(w, req)
	case "/login":
		loginServe(w, req)
	case "/edit":
		editServe(w, req)
	default:
		http.NotFound(w, req)
	}
}

func serve() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(conf.AssetsPath + "/static"))))
	http.HandleFunc("/", serveExact)

	switch conf.ServeMethod {
	case "http":
		log.Fatal(http.ListenAndServe(":" + fmt.Sprintf("%d", conf.HttpPort), nil))
	case "fastcgi":

		// TODO check that it is a socket
		os.Remove(conf.FastcgiPath)

		l, err := net.Listen("unix", conf.FastcgiPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to listen on Fastcgi path '%s': %s.\n", conf.FastcgiPath, err.Error())
			os.Exit(1)
		}
		log.Fatal(fcgi.Serve(l, http.DefaultServeMux))
	}
}
