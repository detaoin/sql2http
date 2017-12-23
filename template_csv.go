package sql2http

import (
	"encoding/csv"
	"fmt"
	"io"
)

// CSVTemplate implements interface Template by writing with csv.Encode the SQL
// query rows to the io.Writer. Each individual SQL query result set is
// separated by an empty line.
type CSVTemplate struct {
	Comma   rune // default ','; e.g. use '\t' for tab-separated values
	UseCRLF bool // use "\r\n" end-of-line character instead of "\n"
}

func (t *CSVTemplate) Execute(wr io.Writer, data interface{}) error {
	resp, ok := data.(*Response)
	if !ok {
		return fmt.Errorf("output csv: only *GetResponse can be passed as data")
	}
	out := csv.NewWriter(wr)
	out.Comma = t.Comma
	out.UseCRLF = t.UseCRLF
	for i, tbl := range resp.Results {
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

func (t *CSVTemplate) ContentType() string {
	return "text/csv"
}
