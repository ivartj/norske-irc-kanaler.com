package web

import (
	_ "code.google.com/p/go-sqlite/go1/sqlite3"
	"database/sql"
	"net/http"
	"html/template"
	"log"
	"fmt"
	"errors"
)

type Site struct {
	dbpath string
	mux map[string]func(Page)
	tpl *template.Template
	fieldData map[string]interface{}
}

func NewTemplate() *template.Template {
	return template.New("").Funcs(template.FuncMap(map[string]interface{}{
		"q" : func(fieldName string) (interface{}, error) {
			return nil, errors.New("The q template dummy function was called")
		},
	}))
}

func NewSite(dbpath string, tpl *template.Template) *Site {

	ctx := &Site{
		dbpath: dbpath,
		mux: map[string]func(Page){},
		tpl: tpl,
		fieldData: map[string]interface{}{},
	}

	return ctx
}

func (ctx *Site) RegisterPage(path string, fn func(Page)) {
	ctx.mux[path] = fn
}

func (ctx *Site) SetField(name string, value interface{}) {
	ctx.fieldData[name] = value
}

func (ctx *Site) SetFieldMap(m map[string]interface{}) {
	for k, v := range m {
		ctx.fieldData[k] = v
	}
}

func (ctx *Site) GetField(name string) (interface{}, error) {
	v, ok := ctx.fieldData[name]
	if !ok {
		return nil, fmt.Errorf("No field data for field '%s'", name)
	}
	return v, nil
}

func (ctx *Site) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	subfn, ok := ctx.mux[req.URL.Path]
	if !ok {
		// TODO: Make own 404 handler
		http.NotFound(w, req)
		return
	}

	subctx := pageNew(ctx, w, req)

	defer func() {
		x := recover()
		if x == nil {
			return
		}
		err := subctx.rollback()
		if err != nil {
			log.Fatalf("Failed to rollback transaction: %s.\n", err.Error())
		}
		// TODO: Serve own error page
		panic(x)
	}()
	subfn(subctx)
	err := subctx.commit()
	if err != nil {
		log.Fatalf("Failed to commit transaction: %s.\n", err.Error())
	}

}

func (ctx *Site) openDB() (*sql.DB, error) {

	c, err := sql.Open("sqlite3", ctx.dbpath)
	if err != nil {
		return nil, fmt.Errorf("Failed to open database connection: %s", err.Error())
	}

	_, err = c.Exec("pragma foreign_keys = 1;")
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("Failed to enable foreign keys in database connection: %s", err.Error())
	}

	return c, nil
}

