package sql2http

import (
	"html/template"
)

var DefaultHTMLTemplate = template.Must(template.New("default.html").Funcs(TemplateFuncs).Parse(defaultHTML))

const defaultHTML = `<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
</head>
<body>
<h1>Results{{with .Request}} for {{.URL.EscapedPath}}{{end}}</h1>
{{range .Results.Slice}}
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
				{{range .Slice -}}
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
