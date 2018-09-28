package sql2http

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

// Executer wraps method Execute. Both text/template.Template
// and html/template.Template implement this interface.
type Executer interface {
	// Execute the template with data, and write the output to wr.
	Execute(wr io.Writer, data interface{}) error
}

// Template is the interface used to format the queried data back to
// the http Response.
//
// It wraps interface Executer with a ContentType method.
type Template interface {
	Executer

	// string to return with the Content-Type Header value.
	ContentType() string
}

// TemplateSet represents a set of templates, stored by file extension.
type TemplateSet struct {
	m sync.RWMutex
	t map[string]Template
}

func (ts *TemplateSet) Register(ext string, t Template) {
	ts.m.Lock()
	if ts.t == nil {
		ts.t = make(map[string]Template)
	}
	ts.t[ext] = t
	ts.m.Unlock()
}

func (ts *TemplateSet) Get(ext string) Template {
	if ts == nil || ts.t == nil {
		return nil
	}
	ts.m.RLock()
	t := ts.t[ext]
	ts.m.RUnlock()
	return t
}

// Clone returns a shallow copy of ts. It allocates a new map, however
// if the Templates associated with ts are pointers, then the returned
// TemplateSet will use the same pointers.
//
// This is typically used to tune part of the templates of
// DefaultTemplateSet. For example:
//
//     t := DefaultTemplateSet.Clone()
//     t[".json"] = myJSONTemplate
//     router.SqlGet(path, queries, t)
func (ts *TemplateSet) Clone() *TemplateSet {
	ts.m.RLock()
	defer ts.m.RUnlock()
	newTS := make(map[string]Template)
	for ext, t := range ts.t {
		newTS[ext] = t
	}
	return &TemplateSet{t: newTS}
}

type executerWithContentType struct {
	Executer
	contentType string
}

func (e *executerWithContentType) ContentType() string { return e.contentType }

// TemplateFromExecuter returns a Template with the given content type string.
func TemplateFromExecuter(t Executer, contentType string) Template {
	return &executerWithContentType{Executer: t, contentType: contentType}
}

// default templates associated with their respective file extension.
// This can be modified before calling the (*Router).SqlXXX methods to
// change the default templates.
// The init functions of the various template/xxx packages each register
// their respective default template.
var DefaultTemplateSet = &TemplateSet{}

var TemplateFuncs = map[string]interface{}{
	"add":   add,
	"mod":   mod,
	"join":  strings.Join,
	"split": strings.Split,
	"tex":   TeXEscaper,
}

func add(x, y int) int { return x + y }
func mod(x, y int) int { return x % y }

func TeXEscapeString(s string) string {
	// "\" replace must be the first one, "\" is used as an escape character!
	s = strings.Replace(s, `\`, `\textbackslash`, -1)
	s = strings.Replace(s, `^`, `\textasciicircum`, -1)
	s = strings.Replace(s, `~`, `\textasciitilde`, -1)
	s = strings.Replace(s, `{`, `\{`, -1)
	s = strings.Replace(s, `}`, `\}`, -1)
	s = strings.Replace(s, `&`, `\&`, -1)
	s = strings.Replace(s, `_`, `\_`, -1)
	s = strings.Replace(s, `#`, `\#`, -1)
	s = strings.Replace(s, `$`, `\$`, -1)
	s = strings.Replace(s, `%`, `\%`, -1)
	return s
}

func TeXEscaper(v interface{}) interface{} {
	switch t := v.(type) {
	case string:
		return TeXEscapeString(t)
	case []string:
		for i := range t {
			t[i] = TeXEscapeString(t[i])
		}
		return t
	}
	return TeXEscapeString(fmt.Sprint(v))
}
