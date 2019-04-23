package tex

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	gopath "path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/detaoin/sql2http/sql2http"
)

func init() {
	sql2http.DefaultTemplateSet.Register(Ext, DefaultTemplate)
}

const (
	Ext = ".tex"
	ContentType = "text/x-tex; charset=utf-8"
)

// Template extends the standard text/template.Template to implement
// interface sql2http.Template (with ContentType method), sets the
// delimiters to `((` & `))`, and adds a new template function `tex`
// to escape strings in (La)TeX documents.
type Template struct {
	*template.Template
}

// ContentType implements interface sql2http.Template.
// It returns the constant string ContentType.
func (t *Template) ContentType() string { return ContentType }

// Lookup returns the template associated with t with given name. This
// method is particularly useful with ParseTree.
//
// If no such template exists, nil is returned.
func (t *Template) Lookup(name string) *Template {
	if t == nil {
		return nil
	}
	tmpl := t.Template.Lookup(name)
	if tmpl == nil {
		return nil
	}
	return &Template{Template: tmpl}
}

// New wraps the standard library text/template.New.
// It is a simple helper function to avoid importing the standard library.
func New(name string) *Template {
	return &Template{
		Template: template.New(name).
			Delims("((", "))").
			Funcs(templateFuncs),
	}
}

// ParseTree creates a new Template, walks recursively the files starting
// at the given directory, and for each file ending with ".tex" parses
// the template definition under the relative file name.
// The returned template is the first matching file.
func ParseTree(dir string) (*Template, error) {
	var t *Template
	err := filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() || filepath.Ext(path) != Ext {
			return nil
		}
		name, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		name = strings.TrimSuffix(name, Ext)
		name = gopath.Clean("/" + filepath.ToSlash(name))
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		s := string(b)
		if t == nil {
			t = New(name)
		}
		var tmpl *template.Template
		if name == t.Name() {
			tmpl = t.Template
		} else {
			tmpl = t.New(name)
		}
		log.Println("found template", name, Ext)
		_, err = tmpl.Parse(s)
		return err
	})
	return t, err
}

func Escape(v interface{}) string {
	var s string
	switch t := v.(type) {
	case string:
		s = t
	case time.Time:
		return t.Format("2006-01-02~15:04:05Z07:00")
	default:
		s = fmt.Sprint(v)
	}
	return EscapeString(s)
}

var escapeChars = []struct{
	from string
	to   string
}{
	{`\`, `\textbackslash`}, // must be first
	{`~`, `\textasciitilde`},
	{`^`, `\textasciicircum`},
	{`&`, `\&`},
	{`%`, `\%`},
	{`$`, `\$`},
	{`#`, `\#`},
	{`_`, `\_`},
	{`#`, `\#`},
	{`{`, `\{`},
	{`}`, `\}`},
}

func EscapeString(s string) string {
	for _, kv := range escapeChars {
		s = strings.Replace(s, kv.from, kv.to, -1)
	}
	return s
}

var templateFuncs = template.FuncMap{
	"tex": Escape,
}

var DefaultTemplate = must(New("_default.tex").Parse(DefaultTeX))

func must(t *template.Template, err error) *Template {
	if err != nil {
		panic(err)
	}
	return &Template{Template: t}
}

// The default tex template. It is valid LaTex (compiled with any latex,
// pdflatex, xelatex, lualatex, or other program).
// It is exported for documentation purposes (can be used as a basis
// for your more elaborate templates).
const DefaultTeX = `\documentclass[a4paper]{article}
\usepackage[utf8]{inputenc}
\begin{document}
\section{Results((with .Request)) for ((.URL.EscapedPath|tex))((end))}
((range .Tables))
\subsection{Table((with .Name)) ((.|tex))((end))}
\begin{tabular}{((range .Header))l((end))}
	\hline
	((range $i, $h := .Header))((if gt $i 0)) & ((end))((tex $h))((end)) \\
	\hline
	\hline
	((range .Rows -))
	((range $i, $v := .Values))((if gt $i 0)) & ((end))((tex $v))((end)) \\
	\hline
	((end))
	\hline
\end{tabular}
((else))
No data available.
((end))
\end{document}
`
