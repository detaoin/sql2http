package sql2http

import (
	"text/template"
)

var DefaultTeXTemplate = template.Must(template.New("default.tex").Delims("((", "))").Funcs(TemplateFuncs).Parse(defaultTeX))

const defaultTeX = `\documentclass[a4paper]{article}
\usepackage[utf8]{inputenc}
\begin{document}
\section{Results((with .Request)) for ((.URL.EscapedPath))((end))}
((range .Results.Slice))
\subsection{Table((if ne .Name "")) ((.Name))((end))}
\begin{tabular}{((range .Header))l((end))}
	\hline
	(( join (tex .Header) " & " )) \\
	\hline
	\hline
	((range .Rows -))
	(( join (tex .Slice) " & " )) \\
	\hline
	((end))
	\hline
\end{tabular}

((else))
No data available.
((end))
\end{document}
`
