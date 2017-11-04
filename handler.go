package sql2http

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"
)

type Handler struct {
	DB *sql.DB

	Pages map[string][]*Pattern

	// default templates for each file name extension
	DefaultTemplates map[string]Template
	// list of templates for specific path patterns, with extension
	FileTemplates map[string]Template

	// logging function to use. If nil, logs are discarded
	Log func(*http.Request, map[string]string, int, error)
}

// discardLog is a noop
func discardLog(*http.Request, map[string]string, int, error) {}

func (h *Handler) findTemplate(pat *Pattern, ext string) (Template, error) {
	if t, ok := h.FileTemplates[pat.String()+ext]; ok {
		return t, nil
	}
	if t, ok := h.DefaultTemplates[ext]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("template not found for %v", pat.String()+ext)
}

func (h *Handler) matchRequest(req *http.Request) (*Pattern, map[string]string) {
	p := req.URL.EscapedPath()
	p = strings.TrimSuffix(p, path.Ext(p)) // trim file extension
	for _, pat := range h.Pages[req.Method] {
		if ok, params := pat.Match(p); ok {
			return pat, params
		}
	}
	return nil, nil
}

func (h *Handler) readFullQuery(queries []SqlQuery, params map[string]string) ([]Table, error) {
	tx, err := h.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	tables := make([]Table, 0, len(queries))
	for _, q := range queries {
		args := make([]interface{}, len(q.Vars))
		for i := range args {
			args[i] = params[q.Vars[i]]
		}
		rows, err := tx.Query(q.Q, args...)
		if err != nil {
			return nil, err
		}
		tbl, err := readTable(rows)
		if err != nil {
			return nil, err
		}
		tbl.Name = q.Name
		tables = append(tables, tbl)
	}

	return tables, tx.Commit()
}

func (h *Handler) runFullExec(queries []SqlQuery, params map[string]string) error {
	tx, err := h.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, q := range queries {
		args := make([]interface{}, len(q.Vars))
		for i := range args {
			args[i] = params[q.Vars[i]]
		}
		if _, err := tx.Exec(q.Q, args...); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (h *Handler) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	if h.Log == nil {
		h.Log = discardLog
	}
	// accept only clean URL paths, redirect otherwise
	p := req.URL.EscapedPath()
	if clean := path.Clean(p); p != clean {
		http.Redirect(wr, req, clean, http.StatusMovedPermanently)
		return
	}

	ext := path.Ext(req.URL.EscapedPath())
	if ext == "" { // default extension is .html
		ext = ".html"
	}

	pat, params := h.matchRequest(req)
	if pat == nil {
		http.NotFound(wr, req)
		h.Log(req, nil, http.StatusNotFound, nil)
		return
	}

	if err := req.ParseForm(); err != nil {
		http.Error(wr, "parsing form error: "+err.Error(), http.StatusInternalServerError)
		h.Log(req, nil, http.StatusInternalServerError, err)
		return
	}

	resp := ResponseFromRequest(req)
	resp.Pattern = pat
	resp.Params = fuseParams(params, req.Form)

	var err error
	switch req.Method {
	case "GET":
		resp.Results, err = h.readFullQuery(pat.Queries, resp.Params)
	case "POST":
		err = h.runFullExec(pat.Queries, resp.Params)
	default:
		http.Error(wr, "Method "+req.Method+" not allowed", http.StatusMethodNotAllowed)
		h.Log(req, resp.Params, http.StatusMethodNotAllowed, err)
		return
	}
	if err != nil {
		http.Error(wr, "query error: "+err.Error(), http.StatusInternalServerError)
		h.Log(req, resp.Params, http.StatusInternalServerError, err)
		return
	}

	tmpl, err := h.findTemplate(pat, ext)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		h.Log(req, resp.Params, http.StatusInternalServerError, err)
		return
	}

	if to, ok := resp.Params["redirect"]; ok {
		http.Redirect(wr, req, to, http.StatusSeeOther)
		h.Log(req, resp.Params, http.StatusSeeOther, nil)
		return
	}
	// TODO: explicitly set mime type to tmpl.Mime()
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, resp); err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		h.Log(req, resp.Params, http.StatusInternalServerError, err)
		return
	}
	_, err = io.Copy(wr, buf)
	h.Log(req, resp.Params, http.StatusOK, err)
}

type SqlQuery struct {
	Name string
	Q    string
	Vars []string
}

type Response struct {
	User      *User             // logged in user, nil if no authentication
	Timestamp time.Time         // timestamp of response
	Version   string            // gowd version string
	Request   *http.Request     `json:"-"` // the http Request used
	Pattern   *Pattern          // the matched URL path pattern
	Params    map[string]string // aggreggated parameters
	Results   Tables
}

// AddToRequest returns a shallow copy of req, with Response r added to its
// Context. The Response can be retrieved with function ResponseFromRequest.
func (r *Response) AddToRequest(req *http.Request) *http.Request {
	ctx := req.Context()
	return req.WithContext(context.WithValue(ctx, keyResponse, r))
}

// ResponseFromRequest returns the Response attached to the request's context.
// The returned Response is always non-nil; it defaults to the zero-value
// Response with the following defaults: Timestamp set to now, Version set to
// the global Version variable, and Request set to req.
func ResponseFromRequest(req *http.Request) *Response {
	ptr := req.Context().Value(keyResponse)
	if ptr == nil {
		return &Response{
			User:      GetUser(req),
			Timestamp: time.Now(),
			Version:   Version,
			Request:   req,
		}
	}
	return ptr.(*Response)
}

type Tables []Table

type Header []string

type Table struct {
	Name   string
	Header []string
	Rows   []Row
}

func (t Tables) Get(name string) (*Table, error) {
	for i, table := range []Table(t) {
		if table.Name == name {
			return &t[i], nil
		}
	}
	return nil, fmt.Errorf("sql2http: table %q not found", name)
}

func (t Tables) Slice() []Table { return []Table(t) }

func (t Tables) MarshalJSON() ([]byte, error) {
	m := make(map[string]Table)
	for _, t := range []Table(t) {
		m[t.Name] = t
	}
	return json.Marshal(m)
}

type Row struct {
	Header []string
	Values []string
}

func (r Row) Get(col string) (string, error) {
	for i, h := range r.Header {
		if h == col {
			return r.Values[i], nil
		}
	}
	return "", fmt.Errorf("sql2http: row element %q not found", col)
}

func (r Row) Slice() []string { return r.Values }

func (r Row) MarshalJSON() ([]byte, error) {
	m := make(map[string]string)
	for i, h := range r.Header {
		m[h] = r.Values[i]
	}
	return json.Marshal(m)
}

func readTable(rows *sql.Rows) (tbl Table, err error) {
	tbl.Header, err = rows.Columns()
	if len(tbl.Header) < 1 {
		return
	}
	dest := make([]interface{}, len(tbl.Header))
	for rows.Next() {
		row := make([]string, len(tbl.Header))
		for i := range dest {
			dest[i] = &row[i]
		}
		if e := rows.Scan(dest...); e == nil {
			tbl.Rows = append(tbl.Rows, Row{tbl.Header, row})
		}
	}
	return tbl, rows.Err()
}

func fuseParams(params map[string]string, form map[string][]string) map[string]string {
	for key, values := range form {
		if _, ok := params[key]; !ok && len(values) > 0 {
			params[key] = values[0]
		}
	}
	return params
}
