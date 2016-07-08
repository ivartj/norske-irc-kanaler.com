package web

import (
	"testing"
	"net/http"
	"strings"
	"os"
)

const tplText = `

{{define "head"}}
<!doctype html>
<html lang="en">
	<head>
		<meta charset="utf-8">
		<title>{{if (q "page-title")}}{{q "page-title"}} - {{q "site-title"}}{{else}}{{q "site-title"}}{{end}}</title>
	</head>
	<body>
{{end}}

{{define "tail"}}
	</body>
</html>
{{end}}

{{define "index"}}
{{template "head"}}
{{q "content"}}
{{template "tail"}}
{{end}}

`

func indexServe(ctx Page) {
	ctx.Set("content", "Lorem ipsum dolor sit amet")
	ctx.ExecuteTemplate("index")
}

type dummyResponseWriter struct{
	header http.Header
}

func newDummyResponseWriter() *dummyResponseWriter {
	return &dummyResponseWriter{
		header: http.Header(map[string][]string{}),
	}
}

func (w *dummyResponseWriter) Header() http.Header {
	return w.header
}

func (w *dummyResponseWriter) Write(bs []byte) (int, error) {
	return os.Stdout.Write(bs)
}

func (w *dummyResponseWriter) WriteHeader(code int) { }

func TestSite(t *testing.T) {
	tpl, err := NewTemplate().Parse(tplText)
	if err != nil {
		t.Fatal(err)
	}
	site := NewSite("./database.db", tpl)
	site.mux["/"] = indexServe
	site.SetFieldMap(map[string]interface{}{
		"page-title" : "",
		"site-title" : "The Greatest Site",
	})

	req, err := http.NewRequest("GET", "/", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	site.ServeHTTP(newDummyResponseWriter(), req)
}

