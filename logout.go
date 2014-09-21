package main

import (
	"net/http"
)

func logoutServe(w http.ResponseWriter, req *http.Request) {
	// TODO: Check if redirect code is appropriate
	if loginAuth(req) {
		loginSessionID = ""
	}
	http.Redirect(w, req, "/", 307)
}
