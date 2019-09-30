package sql2http

import (
	"fmt"
	"io"

	"git.sr.ht/~detaoin/sql2http"
	"github.com/tealeg/xlsx"
)

const Ext = ".xlsx"

func init() {
	sql2http.DefaultTemplateSet.Register(Ext, &XLSXTemplate{})
}

// XLSXTemplate implements interface Template by writing with csv.Encode the
// SQL query rows to the io.Writer. Each individual SQL query result set is
// separated by an empty line.
type XLSXTemplate struct {}

func (t *XLSXTemplate) Execute(wr io.Writer, data interface{}) error {
	resp, ok := data.(*sql2http.Result)
	if !ok {
		return fmt.Errorf("template/xlsx: only *sql2http.Results can be passed as data")
	}
	out := xlsx.NewFile()
	for _, tbl := range resp.Tables {
		name := tbl.Name
		if name == "" {
			name = "default"
		}
		if len(name) > 30 {
			name = name[:27] + "..."
		}
		sh, err := out.AddSheet(name)
		if err != nil {
			return fmt.Errorf("template/xlsx: error creating new sheet: %v", err)
		}
		n := sh.AddRow().WriteSlice(&tbl.Header, -1)
		if n != len(tbl.Header) {
			return fmt.Errorf("template/xlsx[%s]: write header row fail (wrote only %d/%d values)", name, n, len(tbl.Header))
		}
		for i, row := range tbl.Rows {
			n := sh.AddRow().WriteSlice(&row.Values, -1)
			if n != len(row.Values) {
				return fmt.Errorf("template/xlsx[%s:%d]: write data row fail (wrote only %d/%d values)", name, i, n, len(row.Values))
			}
		}
	}
	return out.Write(wr)
}

func (t *XLSXTemplate) ContentType() string {
	// https://blogs.msdn.microsoft.com/vsofficedeveloper/2008/05/08/office-2007-file-format-mime-types-for-http-content-streaming-2/
	return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
}
