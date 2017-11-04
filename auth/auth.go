package auth

import (
	"crypto/sha1"
	"crypto/subtle"
	"fmt"
	"net/http"
	"net/url"

	"github.com/detaoin/sql2http"
	"github.com/gorilla/securecookie"
)

var CookieName = "user-id"

// HTTP auth cookie Max-Age in seconds: default is about 1 month
var MaxAge = 3600 * 24 * 30

var (
	LoginPath  = "/auth/login"
	LogoutPath = "/auth/logout"
)

type UserPass struct {
	Name     string   // copied from sql2http.User
	Fullname string   // copied from sql2http.User
	Tags     []string // copied from sql2http.User

	// The password sha1 hash encoded as a hex string with lowercase
	// letters.
	// For example "adc83b19e793491b1c6ea0fd8b46cd9f32e592fc"
	Pass string
}

func (up *UserPass) User() *sql2http.User {
	return &sql2http.User{
		Name:     up.Name,
		Fullname: up.Fullname,
		Tags:     up.Tags,
	}
}

type authHandler struct {
	users map[string]*UserPass
	h     http.Handler

	s *securecookie.SecureCookie
}

func Handler(h http.Handler, users map[string]*UserPass, hashKey []byte) http.Handler {
	s := securecookie.New(hashKey, nil)
	return &authHandler{users, h, s}
}

func (h *authHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rpath := r.URL.EscapedPath()
	if rpath == LoginPath {
		switch r.Method {
		case "POST":
			h.login(rw, r)
		case "GET":
			h.h.ServeHTTP(rw, r)
		}
		return
	}

	user := h.getUser(rw, r)
	if user == nil {
		v := url.Values{}
		v.Set("redirect", rpath)
		http.Redirect(rw, r, LoginPath+"?"+v.Encode(), http.StatusSeeOther)
		return
	}
	if rpath == LogoutPath {
		h.logout(rw, r)
	}
	r = sql2http.SetUser(r, user)
	h.h.ServeHTTP(rw, r)
}

func (h *authHandler) getUser(rw http.ResponseWriter, r *http.Request) *sql2http.User {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		// TODO: log error
		return nil
	}
	username := ""
	// TODO: log error
	h.s.Decode(CookieName, cookie.Value, &username)
	if userpass := h.users[username]; userpass != nil {
		return userpass.User()
	}
	return nil
}

func (h *authHandler) login(rw http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(rw, "Internal error, try again", http.StatusInternalServerError)
		// TODO: log
		return
	}
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	redirect := r.Form.Get("redirect")
	if redirect == "" {
		redirect = "/"
	}
	user := h.users[username]
	if user == nil {
		notAuthorized(rw, r)
		// TODO: log?
		return
	}
	if !testPasswordSHA1(password, user.Pass) {
		notAuthorized(rw, r)
		// TODO: log?
		return
	}
	encoded, err := h.s.Encode(CookieName, username)
	if err != nil {
		http.Error(rw, "Internal error, try again", http.StatusInternalServerError)
		// TODO: log
		return
	}
	cookie := &http.Cookie{
		Name:   CookieName,
		Value:  encoded,
		Path:   "/",
		MaxAge: MaxAge,
	}
	http.SetCookie(rw, cookie)
	http.Redirect(rw, r, redirect, http.StatusSeeOther)
}

func (h *authHandler) logout(rw http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:   CookieName,
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(rw, cookie)
}

func testPasswordSHA1(password, hash string) bool {
	challenge := fmt.Sprintf("%x", sha1.Sum([]byte(password)))
	return 1 == subtle.ConstantTimeCompare([]byte(challenge), []byte(hash))
}

func notAuthorized(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "username or password incorrect", http.StatusUnauthorized)
}
