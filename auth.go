package sql2http

import (
	"context"
	"net/http"
)

type User struct {
	Name     string
	Fullname string
	Tags     []string
}

func SetUser(req *http.Request, user *User) *http.Request {
	ctx := req.Context()
	return req.WithContext(context.WithValue(ctx, keyUser, user))
}

func GetUser(req *http.Request) *User {
	if v, ok := req.Context().Value(keyUser).(*User); ok {
		return v
	}
	return nil
}
