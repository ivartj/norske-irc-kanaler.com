package web

import (
	"testing"
	"net/http"
	"strings"
	"os"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
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

func indexPage(ctx Page, req *http.Request) {
	ctx.SetField("content", "Lorem ipsum dolor sit amet")
	ctx.ExecuteTemplate("index")
}

func submitPage(ctx Page, req *http.Request) {
	ctx.SetField("content", "Lorem ipsum dolor sit amet")
	_, err := ctx.Exec("insert into test (id) select (max(id) + 1) from test;")
	if err != nil {
		ctx.Fatalf("Error on inserting in table: %s", err.Error())
	}
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
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec("create table test (id integer primary key not null);")
	if err != nil {
		t.Fatal(err)
	}
	tpl, err := NewTemplate().Parse(tplText)
	if err != nil {
		t.Fatal(err)
	}
	site := NewSite(db, tpl)
	site.HandlePage("/", indexPage)
	site.HandlePage("/submit", submitPage)
	site.SetFieldMap(map[string]interface{}{
		"page-title" : "",
		"site-title" : "The Greatest Site",
	})

	req, err := http.NewRequest("GET", "/", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	site.ServeHTTP(newDummyResponseWriter(), req)
	req, err = http.NewRequest("POST", "/submit", strings.NewReader(""))
	site.ServeHTTP(newDummyResponseWriter(), req)
}

