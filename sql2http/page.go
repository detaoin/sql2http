package sql2http

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/julienschmidt/httprouter"
)

// IsolationLevel is the level passed to sql.TxOptions when running
// queries in transactions.
//
// TODO: verify it is supported by all supported database drivers.
var IsolationLevel = sql.LevelSerializable

type page struct {
	pattern   string      // the URL pattern for this page
	templates *TemplateSet // Templates stored by file extension
	queries   []Query

	// fn is set to either runQueries or runExecs depending on the
	// registered http method.
	fn func(context.Context, *sql.DB, *Result) error
	db *sql.DB // the database connection
}

// runQueries runs the list of res.Queries in a single transaction. Then
// the resulting rows are saved in res.Tables.
//
// page.fn is set to runQueries if the page is registered as a GET handler.
func runQueries(ctx context.Context, db *sql.DB, res *Result) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		Isolation: IsolationLevel,
		ReadOnly:  true,
	})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, q := range res.Queries {
		log.Printf("New query: %q\n  %+q\n", q.Q, res.Params)
		params := prepareParams(q, res.Params)
		rows, err := tx.QueryContext(ctx, q.Q, params...)
		if err != nil {
			// TODO: format err?
			return err
		}
		tbl := Table{Name: q.Name}
		if err := readRows(&tbl, rows); err != nil {
			return err
		}
		res.Tables = append(res.Tables, tbl)
	}
	return tx.Commit()
}

// runExecs runs the list of res.Queries in a single transaction.
//
// page.fn is set to runExecs if the page is registered as a POST handler.
func runExecs(ctx context.Context, db *sql.DB, res *Result) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: IsolationLevel})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, q := range res.Queries {
		params := prepareParams(q, res.Params)
		_, err := tx.ExecContext(ctx, q.Q, params...)
		if err != nil {
			// TODO: format err?
			return err
		}
	}
	return tx.Commit()
}

func prepareParams(q Query, p map[string]interface{}) []interface{} {
	args := make([]interface{}, len(q.Params))
	for i, name := range q.Params {
		args[i] = sql.Named(name, p[name])
	}
	return args
}

// readRows reads all data from rows into tbl. It expects a freshly
// returned rows from a Query, and takes care of closing it once done.
func readRows(tbl *Table, rows *sql.Rows) error {
	defer rows.Close()
	var err error
	tbl.Header, err = rows.Columns()
	if err != nil {
		return err
	}
	rowptr := make([]interface{}, len(tbl.Header))
	for rows.Next() {
		row := Row{
			Header: tbl.Header,
			Values: make([]interface{}, len(tbl.Header)),
		}
		for i := range rowptr {
			rowptr[i] = &row.Values[i]
		}
		if err := rows.Scan(rowptr...); err != nil {
			return err
		}
		for i, v := range row.Values {
			if p, ok := v.([]byte); ok {
				row.Values[i] = string(p)
			}
		}
		tbl.Rows = append(tbl.Rows, row)
	}
	return rows.Err()
}

// ServeHTTP implements http.Handler
func (p *page) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	tmpl, err := p.lookupTemplate(req)
	if err != nil {
		// TODO: log error
		http.Error(wr, "no template found", http.StatusNotFound)
		return
	}
	data := &Result{
		Pattern: p.pattern,
		Params:  getParams(req),
		Queries: p.queries,
		Request: Request{req.URL, req.Method, req.Header},
		Time:    time.Now(),
		Version: version,
	}
	if err := p.fn(req.Context(), p.db, data); err != nil {
		http.Error(wr, "error querying the database: " + err.Error(), http.StatusInternalServerError)
		return
	}
	if ct := tmpl.ContentType(); ct != "" {
		wr.Header().Set("Content-Type", ct)
	}
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		http.Error(wr, "error executing the template: " + err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(wr, buf); err != nil {
		log.Println("ERROR", err)
	}
}

func getParams(req *http.Request) map[string]interface{} {
	params := make(map[string]interface{})
	req.ParseForm()
	for k := range req.Form {
		params[k] = req.Form.Get(k)
	}
	for _, kv := range httprouter.ParamsFromContext(req.Context()) {
		params[kv.Key] = kv.Value
	}
	return params
}

func (p *page) lookupTemplate(req *http.Request) (Template, error) {
	ext, _ := req.Context().Value(extKey).(string)
	if ext == "" {
		ext = ".html"
	}
	tmpl := p.templates.Get(ext)
	if tmpl == nil {
		return nil, fmt.Errorf("no template for %q", ext)
	}
	return tmpl, nil
}

type Result struct {
	Pattern string
	Params  map[string]interface{}
	Queries []Query
	Tables  Tables
	Request Request
	Time    time.Time // when the request was made
	Version string    // this package's version
}

type Request struct {
	URL    *url.URL
	Method string
	Header http.Header
}
