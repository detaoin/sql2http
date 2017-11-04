package sql2http

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSONTemplate implements interface Template by writing with json.Encoder the
// full Response data.
type JSONTemplate struct{}

func (t *JSONTemplate) Execute(wr io.Writer, data interface{}) error {
	resp, ok := data.(*Response)
	if !ok {
		return fmt.Errorf("sql2http: output json: only *Response can be passed as data")
	}
	return json.NewEncoder(wr).Encode(resp)
}
