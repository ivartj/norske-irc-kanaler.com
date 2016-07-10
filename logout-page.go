package main

import (
	"net/http"
	"github.com/ivartj/norske-irc-kanaler.com/web"
)

func logoutPage(page web.Page, req *http.Request) {
	// TODO: Check if redirect code is appropriate
	if loginAuth(req) {
		loginSessionID = ""
	}
	http.Redirect(page, req, "/", 307)
}
