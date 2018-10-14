package sql2http

import (
	"context"
	"database/sql"
	"fmt"
	"path"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type Router struct {
	*httprouter.Router
	*sql.DB

	dbdriver string
}

func NewRouter(driver, dataSource string) (*Router, error) {
	db, err := sql.Open(driver, dataSource)
	if err != nil {
		return nil, err
	}
	r := &Router{Router: httprouter.New(), DB: db, dbdriver: driver}
	return r, nil
}

// ServeHTTP wraps the embedded httprouter.Router ServeHTTP to handle
// file extensions.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ext := path.Ext(req.URL.Path)
	// currently httprouter uses
	// req.URL.Path to route the request (see
	// github.com/julienschmidt/httprouter/router.go#L279
	// @commit adbc773).
	// BUG(diego): is modifying req.URL.Path prone to problems?
	req.URL.Path = strings.TrimSuffix(req.URL.Path, ext)
	ctx := req.Context()
	ctx = context.WithValue(ctx, extKey, ext)
	req = req.WithContext(ctx)
	r.Router.ServeHTTP(w, req)
}

// SqlGET registers the path pattern to send the given queries on the
// database upon GET requests.
//
// The path given is the pattern without file extension. When matched
// against a request URL, the URL file extensions is used to find the
// template used (defaulting to ".html").
//
// The list of templates used for the responses is provided with tmpl. It
// defaults to DefaultTemplateSet if nil.
func (r *Router) SqlGET(path string, queries []Query, templates *TemplateSet) {
	for i := range queries {
		bindNamedArgs(r.dbdriver, &queries[i])
	}
	page := &page{
		pattern:   path,
		queries:   queries,
		templates: templates,
		fn:        runQueries,
		db:        r.DB,
	}
	if page.templates == nil {
		page.templates = DefaultTemplateSet
	}
	r.Handler(http.MethodGet, path, page)
}

// SqlPOST registers the path pattern to execute the given queries on the
// database upon POST requests.
//
// The path given is the pattern without file extension. When matched
// against a request URL, the URL file extensions is used to find the
// template used (defaulting to ".html").
//
// The list of templates used for the responses is provided with tmpl. It
// defaults to DefaultTemplateSet if nil.
func (r *Router) SqlPOST(path string, queries []Query, templates *TemplateSet) {
	for i := range queries {
		bindNamedArgs(r.dbdriver, &queries[i])
	}
	page := &page{
		pattern:   path,
		queries:   queries,
		templates: templates,
		fn:        runExecs,
		db:        r.DB,
	}
	if page.templates == nil {
		page.templates = DefaultTemplateSet
	}
	r.Handler(http.MethodPost, path, page)
}

func bindNamedArgs(driver string, q *Query) {
	lexer := lexSQL(q.Q)
	q.Params = q.Params[:0]
	str := strings.Builder{}
	tr := namedArgTranslator(driver)
	o, i := 0, 0
	for tok := range lexer.items {
		if tok.typ == itemIdentifier && tok.val[0] == ':' {
			q.Params = append(q.Params, tok.val[1:])
			str.WriteString(q.Q[o:tok.pos])
			str.WriteString(tr.translate(tok.val, i))
			o = tok.pos + len(tok.val)
			i++
		}
	}
	if o > 0 { // at least 1 named parameter is present
		str.WriteString(q.Q[o:])
		q.Q = str.String()
	}
}

// placeholderType represents the possible placeholders that database
// engines use. The 32 LSBs encoded the rune used as special
// character. The 32 MSBs contain flags.
type placeholderType int64

const (
	placeholderSIMPLE placeholderType = iota
	placeholderNUMBER
	placeholderNAMED
)

func (p placeholderType) typ() placeholderType { return p & 0xffff0000 }

func (p placeholderType) translate(s string, i int) string {
	// fastpath if no translation needed
	if p == placeholderNUMBER|placeholderType(':') {
		return s
	}
	switch p.typ() {
	case placeholderSIMPLE:
		return string(rune(p))
	case placeholderNUMBER:
		return fmt.Sprintf("%c%d", rune(p), i+1)
	case placeholderNAMED:
		return string(rune(p)) + s[1:]
	}
	panic(fmt.Sprintf("sql2http: unknown placeholderType: % x", p))
}

func namedArgTranslator(driver string) placeholderType {
	switch driver {
	case "sqlite3":
		return placeholderNAMED|placeholderType(':')
	case "postgres", "ql":
		return placeholderNUMBER|placeholderType('$')
	case "mysql":
		return placeholderSIMPLE|placeholderType('?')
	case "sqlserver":
		return placeholderNAMED|placeholderType('@')
	default:
		return placeholderSIMPLE|placeholderType('?')
	}
}

type extensionKey struct{}
var extKey = extensionKey{}
