package main

import (
	"net/http"
)

func loginPage(page *page, req *http.Request) {

	page.SetField("redirect", "/")
	page.SetField("success", false)

	switch req.Method {
	case "POST":
		page.SetField("redirect", req.FormValue("redirect"))
		if req.FormValue("password") != page.main.conf.Password() {
			page.AddMessage("Feil passord.")
		} else {
			err := page.main.auth.NewSession()
			if err != nil {
				page.Fatalf("Failed to create new session: %s", err.Error())
			}
			http.SetCookie(page, &http.Cookie{
				Name: "session-id",
				Value: page.main.auth.SessionID(),
				HttpOnly: true,
			})
			page.AddMessage("Innloggin vellykket.")
			page.SetField("success", true)
		}
	case "GET":
		referrer := req.URL.Query().Get("redirect")
		if referrer == "" {
			referrer = req.Referer()
		}
		if referrer != "" {
			page.SetField("redirect", referrer)
		}
	}

	page.ExecuteTemplate("login")
}

