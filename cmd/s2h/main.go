package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/detaoin/sql2http"
	_ "github.com/detaoin/sql2http/template/csv"
	_ "github.com/detaoin/sql2http/template/html"
	_ "github.com/detaoin/sql2http/template/json"
	_ "github.com/detaoin/sql2http/template/tex"
	_ "github.com/detaoin/sql2http/template/xlsx"
)

var addr = flag.String("http", ":8080", "address to expose the http service")

func main() {
	flag.Parse()
	base := filepath.Base(os.Args[0])
	base = base[:len(base)-len(filepath.Ext(base))]
	mux := &sql2http.Router{}
	if err := parseConfig(base, mux); err != nil {
		log.Fatalln(err)
	}
	log.Println("db connected:", mux.Stats())
	log.Fatalln(http.ListenAndServe(*addr, mux))
}
