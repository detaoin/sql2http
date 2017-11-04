package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/user"
	"runtime"

	"github.com/detaoin/sql2http"
)

func initlog() {
	hostname := "Unknown Hostname"
	if hn, err := os.Hostname(); err == nil {
		hostname = hn
	}
	if u, err := user.Current(); err == nil {
		hostname = u.Username + "@" + hostname
	}
	switch mode {
	case mCGI:
		f, err := os.OpenFile(prefix+".log", os.O_WRONLY|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0644)
		if err == nil {
			log.SetOutput(f)
		} else {
			// seems that Apache doesn't like when the CGI script sends to stderr.
			// TODO: verify RFC 3875 (CGI)
			log.SetOutput(ioutil.Discard)
		}
	case mHTTP:
		fallthrough
	default:
		log.SetOutput(os.Stderr)
		log.Printf("Running on %s-%s-%s", hostname, runtime.GOOS, runtime.GOARCH)
	}
}

// fatal handles fatal errors. It is a noop if called with a nil error.
//
// In case of the http server, the message is simply output on stderr, and the
// process exits with a non-zero code.
func fatal(err error) {
	if err == nil {
		return
	}
	log.Output(2, fmt.Sprintln("FATAL", err))
	if mode == mCGI {
		fmt.Printf("Status: 500 Internal Server Error\r\n")
		fmt.Printf("Content-Type: text/plain; charset=utf-8\r\n\r\n")
		fmt.Printf("Fatal error, please contact your administrator")
	}
	os.Exit(1)
}

func logRequest(req *http.Request, params map[string]string, code int, err error) {
	u := sql2http.GetUser(req)
	user := ""
	if u != nil {
		user = fmt.Sprintf(" user:%s", u.Name)
	}
	hosts, e := net.LookupAddr(req.RemoteAddr)
	remote := ""
	if e == nil && len(hosts) > 0 {
		remote = "[" + hosts[0] + "]"
	}
	errstring := ""
	if code >= 400 || err != nil {
		errstring = " ERROR"
	}
	if code >= 400 {
		errstring += ": " + http.StatusText(code)
	}
	if err != nil {
		errstring += ": " + err.Error()
	}
	log.Printf("%v %v %v params:%q%s ip:%s%s%s", req.Method, req.URL, code, params, user, req.RemoteAddr, remote, errstring)
}
