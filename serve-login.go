package main

import (
	"net/http"
	"fmt"
	"crypto/rand"
	"encoding/base64"
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

func (ctx *serveContext) serveLogin(w http.ResponseWriter, req *http.Request) {

	data := struct{
		*serveContext
		Success bool
		Redirect string
	}{
		serveContext: ctx,
	}

	ctx.setPageTitle("Innlogging")

	switch req.Method {
	case "POST":
		data.Redirect = req.FormValue("redirect")
		if req.FormValue("password") != conf.Password {
			ctx.setMessage("Feil passord")
		} else {
			http.SetCookie(w, &http.Cookie{
				Name: "session-id",
				Value: loginNewSessionID(),
				HttpOnly: true,
			})
			ctx.setMessage("Innlogging vellykket.")
			data.Success = true
		}
	case "GET":
		data.Redirect = req.URL.Query().Get("redirect")
	}
		
	err := ctx.executeTemplate(w, "login", &data)
	if err != nil {
		panic(err)
	}
}
