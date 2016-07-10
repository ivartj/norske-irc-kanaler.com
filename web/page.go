package web

import (
	"database/sql"
	"fmt"
	"errors"
	"net/http"
)

type Page interface{
	http.ResponseWriter
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) (*sql.Row)
	ExecuteTemplate(name string)
	SetField(fieldName string, value interface{})
	SetFieldMap(m map[string]interface{})
	GetField(fieldName string) (interface{}, error)
}

type page struct {
	w http.ResponseWriter
	db *sql.DB
	site *Site
	fieldData map[string]interface{}
}

func pageNew(site *Site, w http.ResponseWriter) *page {

	ctx := &page{
		w: w,
		site: site,
		fieldData: map[string]interface{}{},
	}

	return ctx

}

func (ctx *page) Header() http.Header {
	return ctx.w.Header()
}

func (ctx *page) WriteHeader(statusCode int) {
	ctx.w.WriteHeader(statusCode)
}

func (ctx *page) Fatal(args ...interface{}) {
	panic(errors.New(fmt.Sprint(args...)))
}

func (ctx *page) Fatalf(format string, args ...interface{}) {
	panic(fmt.Errorf(format, args...))
}

func (ctx *page) initDb() {
	if ctx.db != nil {
		return
	}

	var err error
	ctx.db, err = ctx.site.openDB()
	if err != nil {
		ctx.Fatal(err.Error())
	}
	_, err = ctx.db.Exec("begin transaction;")
	if err != nil {
		ctx.Fatalf("Failed to initiate database transaction on connection: %s", err.Error())
	}
}

func (ctx *page) Write(bs []byte) (int, error) {
	return ctx.w.Write(bs)
}

func (ctx *page) Exec(query string, args ...interface{}) (sql.Result, error) {
	ctx.initDb()
	return ctx.db.Exec(query, args...)
}

func (ctx *page) Query(query string, args ...interface{}) (*sql.Rows, error) {
	ctx.initDb()
	return ctx.db.Query(query, args...)
}

func (ctx *page) QueryRow(query string, args ...interface{}) *sql.Row{
	ctx.initDb()
	return ctx.db.QueryRow(query, args...)
}

func (ctx *page) commit() error {
	if ctx.db == nil {
		return nil
	}
	_, err := ctx.db.Exec("commit transaction;")
	if err != nil {
		return err
	}
	err = ctx.db.Close()
	if err != nil {
		return err
	}
	return nil
}

func (ctx *page) rollback() error {
	if ctx.db == nil {
		return nil
	}
	_, err := ctx.db.Exec("rollback transaction;")
	if err != nil {
		return err
	}
	err = ctx.db.Close()
	if err != nil {
		return err
	}
	return nil
}

func (ctx *page) ExecuteTemplate(name string) {

	tpl, err := ctx.site.tpl.Clone()
	if err != nil {
		ctx.Fatalf("Failed to clone template: %s", err.Error())
	}

	tpl.Funcs(map[string]interface{}{
		"q" : func(fieldName string) (interface{}, error) {

			v, ok := ctx.fieldData[fieldName]
			if ok {
				return v, nil
			}

			v, ok = ctx.site.fieldData[fieldName]
			if ok {
				return v, nil
			}

			return nil, fmt.Errorf("No field data for field '%s'", fieldName)

		},

	})

	err = tpl.ExecuteTemplate(ctx, name, nil)
	if err != nil {
		ctx.Fatalf("Failed to execute template: %s", err.Error())
	}
}

func (ctx *page) SetField(name string, value interface{}) {
	ctx.fieldData[name] = value
}

func (ctx *page) SetFieldMap(m map[string]interface{}) {
	for k, v := range m {
		ctx.fieldData[k] = v
	}
}

func (ctx *page) GetField(name string) (interface{}, error) {
	v, ok := ctx.fieldData[name]
	if ok {
		return v, nil
	}
	return ctx.site.GetField(name)
}

