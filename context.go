package sql2http

type contextKey int

const (
	keyResponse contextKey = iota
	keyTemplate
	keyUser
)
