package html

import (
	"html/template"
	"io/ioutil"
	"log"
	"os"
	gopath "path"
	"path/filepath"
	"strings"

	"github.com/detaoin/sql2http/sql2http"
)

func init() {
	sql2http.DefaultTemplateSet.Register(Ext, DefaultTemplate)
}

const (
	Ext = ".html"
	ContentType = "text/html; charset=utf-8"
)

// Template extends the standard html/template.Template to implement
// interface sql2http.Template (with ContentType method).
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

// New wraps the standard library html/template.New.
// It is a simple helper function to avoid importing the standard library.
func New(name string) *Template {
	return &Template{Template: template.New(name)}
}

// ParseTree creates a new Template, walks recursively the files starting
// at the given directory, and for each file ending with ".html" parses
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

var DefaultTemplate = must(New("_default.html").Parse(DefaultHTML))

func must(t *template.Template, err error) *Template {
	if err != nil {
		panic(err)
	}
	return &Template{Template: t}
}

// The default html template.
// It is exported for documentation purposes (can be used as a basis
// for your more elaborate templates).
const DefaultHTML = `<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
</head>
<body>
<h1>Results{{with .Request}} for {{.URL.EscapedPath}}{{end}}</h1>
{{range .Tables}}
	<p>{{.Name}}</p>
	<table>
		<thead>
			<tr>
				{{range .Header -}}
				<th title="{{.}}">{{.}}</th>
				{{end}}
			</tr>
		</thead>
		<tbody>
			{{range .Rows}}
			<tr>
				{{range .Values -}}
				<td>{{.}}</td>
				{{- end}}
			</tr>
			{{end}}
		</tbody>
	</table>
{{else}}
	<p>No data available.</p>
{{end}}
<footer>
</footer>
</body>
</html>
`
