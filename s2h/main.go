package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"git.sr.ht/~detaoin/sql2http"
	_ "git.sr.ht/~detaoin/sql2http/template/csv"
	_ "git.sr.ht/~detaoin/sql2http/template/html"
	_ "git.sr.ht/~detaoin/sql2http/template/json"
	_ "git.sr.ht/~detaoin/sql2http/template/tex"
	_ "git.sr.ht/~detaoin/sql2http/template/xlsx"
)

var (
	addr     = ":8080"
	progname = filepath.Base(os.Args[0])
	base     = progname[:len(progname)-len(filepath.Ext(progname))]
)

func main() {
	if env := os.Getenv("S2H"); env != "" {
		base = env
	}
	flag.StringVar(&addr, "http", addr, "address to expose the http service")
	flag.StringVar(&base, "c", base, "configuration files basename")
	flag.Parse()
	mux := &sql2http.Router{}
	if err := parseConfig(base, mux); err != nil {
		log.Fatalln(err)
	}
	log.Println("db connected:", mux.Stats())
	log.Fatalln(http.ListenAndServe(addr, mux))
}
