package sql2http

import (
	"fmt"
	"io"
	"strings"
)

type Template interface {
	Execute(wr io.Writer, data interface{}) error
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
