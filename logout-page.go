package main

import (
	"net/http"
)

func logoutPage(page *page, req *http.Request) {
	if page.main.auth.Nonce() != req.FormValue("nonce") {
		page.AddMessage("Nonce mismatch.")
	} else {
		page.main.auth.Logout()
		page.AddMessage("Logout successful.")
	}

	page.ExecuteTemplate("message")
}
