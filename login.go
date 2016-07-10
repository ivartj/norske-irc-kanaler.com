package main

import (
	"net/http"
	"net/url"
	"fmt"
	"crypto/rand"
	"encoding/base64"
	"github.com/ivartj/norske-irc-kanaler.com/web"
)

// set to "" on logout
var loginSessionID string = ""

func loginNewSessionID() string {
	bytelen := 50
	bytes := make([]byte, bytelen)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(fmt.Errorf("Failed to create random new session ID: %s.\n", err.Error()));
	}
	loginSessionID = base64.URLEncoding.EncodeToString(bytes)
	return loginSessionID
}

func loginAuth(req *http.Request) bool {
	if loginSessionID == "" {
		return false
	}
	c, err := req.Cookie("session-id")
	if err != nil {
		return false
	}
	return loginSessionID == c.Value
}

func loginInit(page web.Page, req *http.Request) {
	if loginAuth(req) {
		page.SetField("admin", true)
		page.SetField("admin-code", loginSessionID)
	}
}

func loginCheck(pageFn func(web.Page, *http.Request)) func(web.Page, *http.Request) {
	return func(page web.Page, req *http.Request) {
		if loginAuth(req) {
			loginInit(page, req)
			pageFn(page, req)
			return
		}
		// TODO: Check if appropriate status code.
		http.Redirect(page, req, "/login?redirect=" + url.QueryEscape(req.URL.Path), 307)
	}
}

func loginPage(page web.Page, req *http.Request) {

	page.SetField("redirect", "/")
	page.SetField("success", false)

	switch req.Method {
	case "POST":
		page.SetField("redirect", req.FormValue("redirect"))
		if req.FormValue("password") != conf.Password {
			utilAddMessage(page, "Feil passord.")
		} else {
			http.SetCookie(page, &http.Cookie{
				Name: "session-id",
				Value: loginNewSessionID(),
				HttpOnly: true,
			})
			utilAddMessage(page, "Innloggin vellykket.")
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

