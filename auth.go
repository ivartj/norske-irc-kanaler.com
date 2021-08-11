package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
)

type auth struct {
	sessionID string
	nonce     string
}

func (a *auth) NewSession() error {
	bytelen := 50
	bytes := make([]byte, bytelen)
	_, err := rand.Read(bytes)
	if err != nil {
		return fmt.Errorf("Failed to create random new session ID: %s.\n", err.Error())
	}
	a.sessionID = base64.URLEncoding.EncodeToString(bytes)

	_, err = rand.Read(bytes)
	if err != nil {
		return fmt.Errorf("Failed to create random new nonce: %s.\n", err.Error())
	}
	a.nonce = base64.URLEncoding.EncodeToString(bytes)

	return nil
}

func (a *auth) Authenticate(req *http.Request) bool {
	if a.sessionID == "" {
		return false
	}
	c, err := req.Cookie("session-id")
	if err != nil {
		return false
	}
	return a.sessionID == c.Value
}

func (a *auth) InitPage(page *page, req *http.Request) {
	if a.Authenticate(req) {
		page.SetField("admin", true)
		page.SetField("auth-nonce", a.nonce)
	}
}

func (a *auth) Guard(pageFn func(*page, *http.Request)) func(*page, *http.Request) {
	return func(page *page, req *http.Request) {
		if a.Authenticate(req) {
			pageFn(page, req)
			return
		}
		// TODO: Check if appropriate status code.
		http.Redirect(page, req, "/login?redirect="+url.QueryEscape(req.URL.Path), 307)
	}
}

func (a *auth) Nonce() string {
	return a.nonce
}

func (a *auth) SessionID() string {
	return a.sessionID
}

func (a *auth) Logout() {
	a.sessionID = ""
	a.nonce = ""
}
