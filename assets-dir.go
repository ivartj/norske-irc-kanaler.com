package main

import (
	"net/http"
)

func assetsDir(page *page, req *http.Request) {
	http.StripPrefix("/static/", http.FileServer(http.Dir(page.main.conf.AssetsPath() + "/static"))).ServeHTTP(page, req)
}

