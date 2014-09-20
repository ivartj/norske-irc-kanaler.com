package main

import (
	"net/http"
	"html/template"
	"log"
	"crypto/rand"
	"encoding/base64"
)

var loginSessionID string = ""

func loginNewSessionID() string {
	bytelen := 50
	bytes := make([]byte, bytelen)
	_, err := rand.Read(bytes)
	if err != nil {
		log.Panicf("Failed to create random new session ID: %s.\n", err.Error());
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

func loginServe(w http.ResponseWriter, req *http.Request) {

	data := struct{
		PageTitle string
		Message string
	}{
		PageTitle: "Innlogging",
	}

	if req.Method == "POST" {
		if req.FormValue("password") != conf.Password {
			data.Message = "Feil passord."
		} else {
			http.SetCookie(w, &http.Cookie{
				Name: "session-id",
				Value: loginNewSessionID(),
				HttpOnly: true,
			})
			data.Message = "Innlogging vellykket."
		}
	}


	tpath := conf.AssetsPath + "/templates.html"
	t, err := template.ParseFiles(tpath)
	if err != nil {
		log.Panicf("Failed to parse template file '%s': %s\n", tpath, err.Error())
	}

	err = t.ExecuteTemplate(w, "login", &data)
}
