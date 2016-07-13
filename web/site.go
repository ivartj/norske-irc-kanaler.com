package web

import (
	"database/sql"
	"net/http"
	"html/template"
	"log"
	"fmt"
	"errors"
	"path"
)

type Site struct {
	db *sql.DB
	pages map[string]func(Page, *http.Request)
	dirs map[string]func(Page, *http.Request)
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

func NewSite(db *sql.DB, tpl *template.Template) *Site {

	ctx := &Site{
		db: db,
		pages: map[string]func(Page, *http.Request){},
		dirs: map[string]func(Page, *http.Request){},
		tpl: tpl,
		fieldData: map[string]interface{}{},
	}

	return ctx
}

func (ctx *Site) HandlePage(path string, fn func(Page, *http.Request)) {
	ctx.pages[path] = fn
}

func (ctx *Site) HandleDir(path string, fn func(Page, *http.Request)) {
	ctx.dirs[path] = fn
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

func (ctx *Site) getHandler(p string) (func(Page, *http.Request), bool) {

	h, ok := ctx.pages[p]
	if ok {
		return h, true
	}

	for _, v := range []map[string]func(Page, *http.Request){ctx.pages, ctx.dirs} {
		h, ok = v[p + "/"]
		if ok {
			return func(page Page, req *http.Request) {
				http.Redirect(page, req, p + "/", http.StatusMovedPermanently)
			}, true
		}
	}

	for p != "/" {
		p = path.Dir(p)
		h, ok = ctx.dirs[p + "/"]
		if ok {
			return h, true
		}
	}

	return nil, false
}

func (ctx *Site) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	subfn, ok := ctx.getHandler(req.URL.Path)
	if !ok {
		// TODO: Make own 404 handler
		http.NotFound(w, req)
		return
	}

	subctx := pageNew(ctx, w)

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
	subfn(subctx, req)
	err := subctx.commit()
	if err != nil {
		log.Fatalf("Failed to commit transaction: %s.\n", err.Error())
	}

}

func (ctx *Site) Begin() (*sql.Tx, error) {
	return ctx.db.Begin()
}

