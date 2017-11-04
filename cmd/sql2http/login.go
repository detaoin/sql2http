package main

import (
	"io"
	"log"
	"net/http"

	"github.com/detaoin/sql2http"
	"github.com/detaoin/sql2http/auth"
)

const defaultLoginPage = `<!doctype html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
</head>
<body>
<form method="POST" action="/auth/login">
username: <input type="text" id="username" name="username" required /><br />
password: <input type="password" id="password" name="password" required /><br />
<button type="submit" />
</form>
</body>
</html>
`

const defaultLogoutPage = `<!doctype html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
</head>
<body>
You have been successfully logged out.
</body>
</html>
`

func loginHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}
	if tmpl := handler.FileTemplates[auth.LoginPath+".html"]; tmpl != nil {
		resp := sql2http.ResponseFromRequest(r)
		// TODO: error handling
		tmpl.Execute(w, resp)
		return
	}
	log.Printf("WARN template %q not found in %v", auth.LoginPath+".html")
	io.WriteString(w, defaultLoginPage)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}
	if tmpl := handler.FileTemplates[auth.LogoutPath+".html"]; tmpl != nil {
		resp := sql2http.ResponseFromRequest(r)
		// TODO: error handling
		tmpl.Execute(w, resp)
		return
	}
	log.Printf("WARN template %q not found", auth.LogoutPath+".html")
	io.WriteString(w, defaultLogoutPage)
}
