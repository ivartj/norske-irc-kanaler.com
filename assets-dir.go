package main

import (
	"github.com/ivartj/norske-irc-kanaler.com/web"
	"net/http"
)

func assetsDir(page web.Page, req *http.Request) {
	http.StripPrefix("/static/", http.FileServer(http.Dir(conf.AssetsPath + "/static"))).ServeHTTP(page, req)
}

