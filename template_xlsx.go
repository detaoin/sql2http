package sql2http

import (
	"fmt"
	"io"

	"github.com/tealeg/xlsx"
)

// XLSXTemplate implements interface Template by writing with csv.Encode the
// SQL query rows to the io.Writer. Each individual SQL query result set is
// separated by an empty line.
type XLSXTemplate struct {
	Comma   rune // default ','; e.g. use '\t' for tab-separated values
	UseCRLF bool // use "\r\n" end-of-line character instead of "\n"
}

func (t *XLSXTemplate) Execute(wr io.Writer, data interface{}) error {
	resp, ok := data.(*Response)
	if !ok {
		return fmt.Errorf("output xlsx: only *Response can be passed as data")
	}
	out := xlsx.NewFile()
	for _, tbl := range resp.Results {
		name := tbl.Name
		if name == "" {
			name = "default"
		}
		sh, err := out.AddSheet(name)
		if err != nil {
			return fmt.Errorf("output xlsx: error creating new sheet: %v", err)
		}
		n := sh.AddRow().WriteSlice(&tbl.Header, -1)
		if n != len(tbl.Header) {
			return fmt.Errorf("output xlsx[%s]: write header row fail (%d/%d)", name, n, len(tbl.Header))
		}
		for i, row := range tbl.Rows {
			n := sh.AddRow().WriteSlice(&row.Values, -1)
			if n != len(row.Values) {
				return fmt.Errorf("output xlsx[%s:%d]: write data row fail (%d/%d)", name, i, n, len(row.Values))
			}
		}
	}
	return out.Write(wr)
}

func (t *XLSXTemplate) ContentType() string {
	// https://blogs.msdn.microsoft.com/vsofficedeveloper/2008/05/08/office-2007-file-format-mime-types-for-http-content-streaming-2/
	return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
}
