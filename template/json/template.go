package json

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/detaoin/sql2http"
)

const Ext = ".json"

func init() {
	sql2http.DefaultTemplateSet.Register(Ext, &Template{})
}

// Template implements interface sql2http.Template by writing with json.Encoder the
// full Response data.
type Template struct{}

func (t *Template) Execute(wr io.Writer, data interface{}) error {
	resp, ok := data.(*sql2http.Result)
	if !ok {
		return fmt.Errorf("sql2http: template/json: only *sql2http.Result is valid data")
	}
	return json.NewEncoder(wr).Encode(resp)
}

func (t *Template) ContentType() string {
	return "application/json; charset=utf-8"
}
