package main

import (
	"database/sql"
	"fmt"
	htmltmpl "html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/detaoin/sql2http"
	"github.com/detaoin/sql2http/auth"
	"github.com/ghodss/yaml"
)

type Config struct {
	DB struct {
		Driver  string
		Options string
	}
	Auth *struct {
		Users  []*auth.UserPass
		Secret string
	}

	// Get holds all the GET requests.
	Get map[string][]sql2http.SqlQuery
	// Get holds all the POST requests.
	Post map[string][]sql2http.SqlQuery

	// list of templates ordered by request path. The extension is included.
	Templates map[string]string
}

// Parse parses the config files with given prefix.
//
//     - <prefix>.yaml
//     - <prefix>.auth
func (c *Config) Parse(prefix string) error {
	if err := yamlDecodeFile(prefix+".yaml", &c); err != nil {
		return err
	}
	if err := c.parseTemplates(prefix + ".templates"); err != nil {
		return err
	}
	if config.Auth != nil {
		if len(config.Auth.Secret) <= 0 {
			return fmt.Errorf("config: auth.secret must be defined")
		}
	}
	return nil
}

func (c *Config) Handler() (*sql2http.Handler, error) {
	db, err := sql.Open(c.DB.Driver, c.DB.Options)
	if err != nil {
		return nil, err
	}
	pages := make(map[string][]*sql2http.Pattern)

	getPages := make([]*sql2http.Pattern, 0)
	for uri, q := range c.Get {
		pat := sql2http.ParsePattern(uri)
		pat.Queries = q
		getPages = append(getPages, pat)
	}
	pages["GET"] = getPages

	postPages := make([]*sql2http.Pattern, 0)
	for uri, q := range c.Post {
		pat := sql2http.ParsePattern(uri)
		pat.Queries = q
		postPages = append(postPages, pat)
	}
	pages["POST"] = postPages

	d := sql2http.DefaultTemplates
	fileTemplates := make(map[string]sql2http.Template)
	for uri, path := range c.Templates {
		ext := filepath.Ext(path)
		var tmpl sql2http.Template
		switch ext {
		case ".html":
			var t *htmltmpl.Template
			t, err = htmltmpl.New(filepath.Base(path)).Funcs(sql2http.TemplateFuncs).ParseFiles(path)
			if t2, e := t.ParseGlob(filepath.Join(prefix+".templates", "_*.html")); e == nil {
				t = t2
			}
			tmpl = t
		case ".tex":
			var t *template.Template
			t, err = template.New(filepath.Base(path)).Delims("(", ")").Funcs(sql2http.TemplateFuncs).ParseFiles(path)
			if t2, e := t.ParseGlob(filepath.Join(prefix+".templates", "_*.tex")); e == nil {
				t = t2
			}
			tmpl = t
		default:
			tmpl, err = template.ParseFiles(path)
		}
		if err != nil {
			return nil, err
		}
		if len(ext) != 0 && uri[:len(uri)-len(ext)] == "/_default" {
			d[ext] = tmpl
		} else {
			fileTemplates[uri] = tmpl
		}
	}

	return &sql2http.Handler{
		DB:               db,
		Pages:            pages,
		DefaultTemplates: d,
		FileTemplates:    fileTemplates,
		Log:              logRequest,
	}, nil
}

func (c *Config) parseTemplates(root string) error {
	c.Templates = make(map[string]string)
	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		uri := filepath.ToSlash(filepath.Clean(path[len(root):]))
		c.Templates[uri] = path
		return nil
	}
	return filepath.Walk(root, walk)
}

func yamlDecodeFile(file string, o interface{}) error {
	p, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(p, o)
}
