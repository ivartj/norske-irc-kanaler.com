package main

import (
	"net/http"
)

func (ctx *serveContext) serveError(w http.ResponseWriter, req *http.Request, msg string) {
	data := struct{
		*serveContext
		Admin bool
		Message string
	}{
		serveContext: ctx,
		Message: msg,
	}

	ctx.setPageTitle("Feilmelding")

	err := ctx.executeTemplate(w, "error", &data)
	if err != nil {
		panic(err)
	}
}
