package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/detaoin/sql2http"
	"github.com/detaoin/sql2http/template/html"
	"github.com/detaoin/sql2http/template/tex"
)

func parseConfig(base string, mux *sql2http.Router) error {
	file := base + ".conf"
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	p := &parser{Router: mux}
	p.tmplHtml, _ = html.ParseTree(base + ".templates")
	p.tmplTex, _ = tex.ParseTree(base + ".templates")
	p.tmpls = sql2http.DefaultTemplateSet
	if t := p.tmplHtml.Lookup("_default"); t != nil {
		p.tmpls.Register(html.Ext, t)
	}
	if t := p.tmplTex.Lookup("_default"); t != nil {
		p.tmpls.Register(html.Ext, t)
	}
	lines := bufio.NewScanner(f)
	i := 0
	for lines.Scan() {
		i++
		if i == 1 {
			m, err := parseDBLine(lines.Text())
			if err != nil {
				return fmt.Errorf("%s:%d: %v", file, i, err)
			}
			*mux = *m
			continue
		}

		if err := p.next(lines.Text()); err != nil {
			return fmt.Errorf("%s:%d: %v", file, i, err)
		}
	}
	if err := p.next(""); err != nil {
		return fmt.Errorf("%s:%d: %v", file, i, err)
	}
	return nil
}

func parseDBLine(line string) (*sql2http.Router, error) {
	toks := strings.SplitN(line, " ", 2)
	options := ""
	if len(toks) == 2 {
		options = strings.TrimSpace(toks[1])
	}
	return sql2http.NewRouter(strings.TrimSpace(toks[0]), options)
}

type parser struct {
	*sql2http.Router

	path    string // if empty, it means we expect one!
	method  string
	queries []sql2http.Query
	query   strings.Builder // current query, possibly on multiple lines

	tmpls    *sql2http.TemplateSet
	tmplHtml *html.Template
	tmplTex  *tex.Template
}

func (p *parser) next(line string) error {
	trimline := strings.TrimSpace(line)
	firstchar, _ := utf8.DecodeRuneInString(line)
	switch {
	case p.path == "" && trimline == "":
		// nothing
	case trimline == "":
		if err := p.register(); err != nil {
			return err
		}
		p.path = ""
		p.method = ""
		p.queries = nil
	case strings.HasPrefix(line, "GET"):
		p.method = "GET"
		fallthrough
	case strings.HasPrefix(line, "POST"):
		if p.method == "" {
			p.method = "POST"
		}
		toks := strings.Fields(trimline)
		if len(toks) != 2 || toks[0] != p.method {
			return fmt.Errorf("invalid path line")
		}
		p.path = toks[1]
	case !unicode.IsSpace(firstchar): // new query
		if p.query.Len() > 0 {
			p.queries[len(p.queries)-1].Q = p.query.String()
			p.query.Reset()
		}
		toks := strings.SplitN(trimline, ":", 2)
		if len(toks) != 2 {
			return fmt.Errorf("missing ':'")
		}
		p.queries = append(p.queries, sql2http.Query{Name: strings.TrimSpace(toks[0])})
		p.query.WriteString(strings.TrimSpace(toks[1]))
	default:
		p.query.WriteByte(' ')
		p.query.WriteString(trimline)
	}
	return nil
}

func (p *parser) register() error {
	if p.path == "" {
		return nil
	}
	if p.query.Len() > 0 {
		// need to flush pending query string
		p.queries[len(p.queries)-1].Q = p.query.String()
		p.query.Reset()
	}
	ts := p.tmpls
	log.Println(p.path, ":", ts.Get(html.Ext).(*html.Template).Name())
	name := strings.TrimPrefix(p.path, "/")
	if th, tt := p.tmplHtml.Lookup(name), p.tmplTex.Lookup(name); th != nil || tt != nil {
		log.Println(name, "customize templates:")
		ts = ts.Clone()
		if th != nil {
			log.Println(" ", html.Ext)
			ts.Register(html.Ext, th)
			log.Println("default:", p.tmpls.Get(html.Ext).(*html.Template).Name())
			log.Println(p.path+":", ts.Get(html.Ext).(*html.Template).Name())
		}
		if tt != nil {
			log.Println(" ", tex.Ext)
			ts.Register(tex.Ext, tt)
		}
	}
	switch p.method {
	case "GET":
		log.Printf("GET  %q %+q\n", p.path, p.queries)
		p.SqlGET(p.path, p.queries, ts)
		log.Println(p.path, ":", ts.Get(html.Ext).(*html.Template).Name())
	case "POST":
		log.Printf("POST %q %+q\n", p.path, p.queries)
		p.SqlPOST(p.path, p.queries, ts)
		log.Println(p.path, ":", ts.Get(html.Ext).(*html.Template).Name())
	default:
		return fmt.Errorf("invalid HTTP method %v", p.method)
	}
	return nil
}
