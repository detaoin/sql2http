package sql2http

import (
	"reflect"
	"testing"
)

func TestParsePattern(t *testing.T) {
	testcases := []struct {
		in   string
		want *Pattern
	}{
		{
			"..",
			&Pattern{},
		},
		{
			".//../",
			&Pattern{},
		},
		{
			"",
			&Pattern{},
		},
		{
			"/foo/@bar/",
			&Pattern{
				s:       "foo/@bar",
				anchors: []token{{0, "foo"}},
				vars:    []token{{1, "bar"}},
			},
		},
	}

	for _, tc := range testcases {
		got := ParsePattern(tc.in)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("parse %q; want %+q, got %+q", tc.in, tc.want, got)
		}
	}
}
