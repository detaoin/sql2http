package sql2http

import (
	"fmt"
	"io"
	"strings"
)

// Executer wraps method Execute. Both text/template.Template
// and html/template.Template implement this interface.
type Executer interface {
	// Execute the template with data, and write the output to wr.
	Execute(wr io.Writer, data interface{}) error
}

type Template interface {
	Executer
	// string to return with the Content-Type Header value.
	ContentType() string
}

type executerWithContentType struct {
	e           Executer
	contentType string
}

func (e *executerWithContentType) Execute(wr io.Writer, data interface{}) error { return e.e.Execute(wr, data) }

func (e *executerWithContentType) ContentType() string { return e.contentType }

func WrapStandardTemplate(t Executer, contentType string) Template {
	return &executerWithContentType{t, contentType}
}

// default templates associated with their respective file extension.
var DefaultTemplates = map[string]Template{
	".csv":  &CSVTemplate{Comma: ','},
	".json": &JSONTemplate{},
	".tsv":  &CSVTemplate{Comma: '\t'},
	".tex":  DefaultTeXTemplate,
	".html": DefaultHTMLTemplate,
	".xlsx": &XLSXTemplate{},
}

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
	s = strings.Replace(s, `{`, `\{`, -1)
	s = strings.Replace(s, `}`, `\}`, -1)
	s = strings.Replace(s, `&`, `\&`, -1)
	s = strings.Replace(s, `_`, `\_`, -1)
	s = strings.Replace(s, `^`, `\^`, -1)
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
