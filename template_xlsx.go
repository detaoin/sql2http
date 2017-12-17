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
	now := resp.Timestamp.Format("2006-01-02 15:04:05")
	for _, tbl := range resp.Results {
		sh, err := out.AddSheet(fmt.Sprintf("%v (%v)", tbl.Name, now))
		if err != nil {
			return fmt.Errorf("output xlsx: error creating new sheet: %v", err)
		}
		sh.AddRow().WriteSlice(&tbl.Header, 0)
		for _, row := range tbl.Rows {
			sh.AddRow().WriteSlice(&row.Values, 0)
		}
	}
	return out.Write(wr)
}
