package main

import (
	"fmt"
	"io/ioutil"

	"git.sr.ht/~detaoin/sql2http"
	"gopkg.in/yaml.v2"
)

type yamlConf struct {
	Db struct {
		Driver  string
		Options string
	}
	Pages []struct {
		Pattern string
		Method  string
		Queries yaml.MapSlice
	}
}

func parseYAML(file string, mux *sql2http.Router, templates *Templates) error {
	conf := yamlConf{}
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(b, &conf); err != nil {
		return err
	}
	m, err := sql2http.NewRouter(conf.Db.Driver, conf.Db.Options)
	if err != nil {
		return err
	}
	*mux = *m
	for _, page := range conf.Pages {
		if page.Pattern == "" {
			return fmt.Errorf("%v: pages.pattern must be non-empty; found %q", file, page.Pattern)
		}
		queries := make([]sql2http.Query, len(page.Queries))
		for i, q := range page.Queries {
			queries[i].Name = fmt.Sprint(q.Key)
			switch v := q.Value.(type) {
			case string:
				queries[i].Q = v
			case []byte:
				queries[i].Q = string(v)
			default:
				return fmt.Errorf("%v:%v:%v: invalid SQL query", file, page.Pattern, queries[i].Name)
			}
		}
		switch page.Method {
		case "GET":
			mux.SqlGET(page.Pattern, queries, templates.GetTemplateSet(page.Pattern))
		case "POST":
			mux.SqlPOST(page.Pattern, queries, templates.GetTemplateSet(page.Pattern))
		default:
			return fmt.Errorf("%v:%v: invalid method %q", file, page.Pattern, page.Method)
		}
	}
	return nil
}
