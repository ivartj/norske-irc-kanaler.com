package main

import (
	"net/http"
	"html/template"
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

func loginServe(w http.ResponseWriter, req *http.Request) {

	data := struct{
		serveCommon
		PageTitle string
		Message string
		Success bool
		Redirect string
	}{
		serveCommon: serveCommonData(req),
		PageTitle: "Innlogging",
	}

	switch req.Method {
	case "POST":
		data.Redirect = req.FormValue("redirect")
		if req.FormValue("password") != conf.Password {
			data.Message = "Feil passord."
		} else {
			http.SetCookie(w, &http.Cookie{
				Name: "session-id",
				Value: loginNewSessionID(),
				HttpOnly: true,
			})
			data.Message = "Innlogging vellykket."
			data.Success = true
		}
	case "GET":
		data.Redirect = req.URL.Query().Get("redirect")
	}
		


	tpath := conf.AssetsPath + "/templates.html"
	t, err := template.ParseFiles(tpath)
	if err != nil {
		panic(fmt.Errorf("Failed to parse template file '%s': %s\n", tpath, err.Error()))
	}

	err = t.ExecuteTemplate(w, "login", &data)
	if err != nil {
		panic(fmt.Errorf("Failed to execute template file '%s': %s\n", tpath, err.Error()))
	}
}
