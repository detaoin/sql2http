package sql2http

import (
	"context"
	"path"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/jmoiron/sqlx"
)

type Router struct {
	*httprouter.Router
	*sqlx.DB
}

func NewRouter(driver, dataSource string) (*Router, error) {
	db, err := sqlx.Open(driver, dataSource)
	if err != nil {
		return nil, err
	}
	r := &Router{Router: httprouter.New(), DB: db}
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

type extensionKey struct{}
var extKey = extensionKey{}
