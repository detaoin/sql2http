package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/cgi"
	"os"
	"path/filepath"

	"github.com/detaoin/sql2http"
	"github.com/detaoin/sql2http/auth"
)

// command-line flags
var (
	fHTTP = flag.String("http", ":8888", "launch a local http server and listen on `address`")
)

type Mode int

// modes
const (
	mUnknown Mode = iota
	mHTTP
	mCGI
)

var Version = "development"

var (
	// prefix is the filename prefix used to find configuration files. By
	// default it is this executable filename with all filename extensions
	// stripped.
	prefix string

	// HTTP or CGI mode
	mode Mode

	// holds the parsed configuration data.
	config = &Config{}

	// the sql2http Handler
	handler *sql2http.Handler
)

func main() {
	flag.Parse()
	os.Chdir(filepath.Dir(os.Args[0]))
	prefix = filepath.Base(os.Args[0])
	ext := filepath.Ext(prefix)
	prefix = prefix[:len(prefix)-len(ext)]
	mode = mHTTP
	req, err := cgi.Request()
	if err == nil && req != nil {
		mode = mCGI
	}
	initlog()

	fatal(config.Parse(prefix))

	handler, err = config.Handler()
	fatal(err)
	mux := http.NewServeMux()
	mux.Handle("/", handler)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(prefix+".static"))))
	website := http.Handler(mux)
	if config.Auth != nil { // wrap mux with authentication stuff
		mux.HandleFunc("/auth/login", loginHandler)
		mux.HandleFunc("/auth/logout", logoutHandler)
		users := make(map[string]*auth.UserPass)
		for _, u := range config.Auth.Users {
			users[u.Name] = u
		}
		website = auth.Handler(mux, users, []byte(config.Auth.Secret))
	}

	switch mode {
	case mHTTP:
		log.Println("now listening on", *fHTTP)
		fatal(http.ListenAndServe(*fHTTP, website))
	case mCGI:
		fatal(ServeCGI(website, req))
	default:
		fatal(fmt.Errorf("unhandled mode: %v", mode))
	}
	log.Println("clean exit")
}
