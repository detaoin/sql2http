package sql2http

import (
	"encoding/json"
	"path"
	"strings"
)

// Pattern represents a URL path matching pattern. Relative paths are
// considered as implicitly rooted, and are interpreted as a sequence of
// slash-separated tokens.  Trailing slashes are ignored, and consecutive
// slashes are evaluated as a single slash; the empty string is not a
// possible/valid token. Dot, and dot-dot are simplified using path.Clean
// before interpreting the path. These properties are implemented using
// path.Clean.
//
// If a token needs to contain a slash, it has to be URL-encoded.
//
// The zero-value Pattern corresponds to the root path "/" only.
type Pattern struct {
	s       string
	anchors []token // fixed tokens which must match
	vars    []token // variable tokens

	Queries []SqlQuery // list of queries associated with this pattern
}

// token represents a single token in a slash-separated path.
type token struct {
	i int    // index in the path, zero-based.
	v string // token value.
}

// ParsePattern returns a Pattern variable corresponding to the encoded pattern
// string s. path.Clean is called before parsing the string.
func ParsePattern(s string) *Pattern {
	s = cleanPath(s)
	if s == "" {
		return &Pattern{}
	}
	pat := &Pattern{s: s}
	toks := strings.Split(s, "/")
	for i, t := range toks {
		if strings.HasPrefix(t, "@") {
			pat.vars = append(pat.vars, token{i, t[1:]})
		} else {
			pat.anchors = append(pat.anchors, token{i, t})
		}
	}
	return pat
}

func cleanPath(s string) string {
	return strings.TrimPrefix(path.Clean("/"+s), "/")
}

// String returns the canonical path which parses to the given pattern.
func (p *Pattern) String() string {
	return "/" + p.s
}

func (p *Pattern) MarshalJSON() ([]byte, error) {
	b := []byte(`{"Path": `)
	buf, err := json.Marshal(p.String())
	if err != nil {
		return nil, err
	}
	b = append(b, buf...)
	b = append(b, []byte(`, "Queries": `)...)
	buf, err = json.Marshal(p.Queries)
	if err != nil {
		return nil, err
	}
	buf = append(b, buf...)
	buf = append(buf, '}')
	return buf, err
}

// Len returns the number of tokens represented by this Pattern.
func (p *Pattern) Len() int {
	if p == nil {
		return 0
	}
	return len(p.anchors) + len(p.vars)
}

// Match reports whether the Pattern matches the string s. If and only if it
// matches, does it return a non-nil map of the variable token values.
//
// Example: calling Match on Pattern created from "/foo/@bar" with path "/foo"
// returns (false, nil). However if called with path "/foo/baz" it returns
// (true, {"bar": "baz"}).
func (p *Pattern) Match(s string) (bool, map[string]string) {
	s = cleanPath(s)
	if s == "" {
		if p.Len() == 0 {
			return true, make(map[string]string)
		}
		return false, nil
	}
	toks := strings.Split(s, "/")
	if p.Len() != len(toks) {
		return false, nil
	}
	for _, a := range p.anchors {
		if toks[a.i] != a.v {
			return false, nil
		}
	}

	// Match is successful, start allocating memory
	params := make(map[string]string)
	for _, v := range p.vars {
		params[v.v] = toks[v.i]
	}
	return true, params
}
