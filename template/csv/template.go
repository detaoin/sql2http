package csv

import (
	"encoding/csv"
	"fmt"
	"io"

	"git.sr.ht/~detaoin/sql2http"
)

func init() {
	sql2http.DefaultTemplateSet.Register(".csv", &Template{Comma: ','})
	sql2http.DefaultTemplateSet.Register(".tsv", &Template{Comma: '\t'})
}

// Template implements interface sql2http.Template by writing with
// csv.Encode the SQL query rows to the io.Writer. Each individual SQL
// query result set is separated by an empty line.
type Template struct {
	Comma   rune // default ','; e.g. use '\t' for tab-separated values
	UseCRLF bool // use "\r\n" end-of-line character instead of "\n"
}

func (t *Template) Execute(wr io.Writer, data interface{}) error {
	resp, ok := data.(*sql2http.Result)
	if !ok {
		return fmt.Errorf("template/csv: only *sql2http.Result can be passed as data")
	}
	out := csv.NewWriter(wr)
	out.Comma = t.Comma
	out.UseCRLF = t.UseCRLF
	for i, tbl := range resp.Tables {
		if i > 0 { // separate tables with an empty line
			out.Write([]string{})
		}
		vals := make([]string, len(tbl.Header))
		out.Write(tbl.Header)
		for _, row := range tbl.Rows {
			for i := range row.Values {
				vals[i] = fmt.Sprint(row.Values[i])
			}
			out.Write(vals)
		}
	}
	out.Flush()
	return out.Error()
}

func (t *Template) ContentType() string {
	switch t.Comma {
	case ',':
		return "text/csv"
	case '\t':
		return "text/tab-separated-values"
	default:
		return "text/plain"
	}
}
