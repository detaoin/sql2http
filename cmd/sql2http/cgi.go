package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
)

// The following are copied from net/http/cgi/child.go

func ServeCGI(h http.Handler, r *http.Request) error {
	rw := &response{
		req:    r,
		header: make(http.Header),
		bufw:   bufio.NewWriter(os.Stdout),
	}
	h.ServeHTTP(rw, r)
	rw.Write(nil) // make sure a response is sent
	if err := rw.bufw.Flush(); err != nil {
		return err
	}
	return nil
}

type response struct {
	req        *http.Request
	header     http.Header
	bufw       *bufio.Writer
	headerSent bool
}

func (r *response) Flush() {
	r.bufw.Flush()
}

func (r *response) Header() http.Header {
	return r.header
}

func (r *response) Write(p []byte) (n int, err error) {
	if !r.headerSent {
		r.WriteHeader(http.StatusOK)
	}
	return r.bufw.Write(p)
}

func (r *response) WriteHeader(code int) {
	if r.headerSent {
		// Note: explicitly using Stderr, as Stdout is our HTTP output.
		// TODO: write to the cgi log
		fmt.Fprintf(os.Stderr, "CGI attempted to write header twice on request for %s", r.req.URL)
		return
	}
	r.headerSent = true
	fmt.Fprintf(r.bufw, "Status: %d %s\r\n", code, http.StatusText(code))

	// Set a default Content-Type
	// TODO: sniff Content-Type if not explicitly set
	if _, hasType := r.header["Content-Type"]; !hasType {
		r.header.Add("Content-Type", "text/html; charset=utf-8")
	}

	r.header.Write(r.bufw)
	r.bufw.WriteString("\r\n")
	r.bufw.Flush()
}
