package main

import (
	"github.com/detaoin/sql2http/sql2http"
	"github.com/detaoin/sql2http/template/html"
	"github.com/detaoin/sql2http/template/tex"
)

type Templates struct {
	*sql2http.TemplateSet
	htmlTree *html.Template
	texTree  *tex.Template
}

func parseTemplates(dir string) *Templates {
	templates := &Templates{TemplateSet: sql2http.DefaultTemplateSet}
	templates.htmlTree, _ = html.ParseTree(dir)
	templates.texTree, _ = tex.ParseTree(dir)
	if t := templates.htmlTree.Lookup("/_default"); t != nil {
		templates.Register(html.Ext, t)
	}
	if t := templates.texTree.Lookup("/_default"); t != nil {
		templates.Register(tex.Ext, t)
	}
	return templates
}

func (tmpls *Templates) GetTemplateSet(pattern string) *sql2http.TemplateSet {
	ts := tmpls.TemplateSet
	if t := tmpls.htmlTree.Lookup(pattern); t != nil {
		ts = ts.Clone()
		ts.Register(html.Ext, t)
	}
	if t := tmpls.texTree.Lookup(pattern); t != nil {
		if ts == tmpls.TemplateSet {
			ts = ts.Clone()
		}
		ts.Register(tex.Ext, t)
	}
	return ts
}
