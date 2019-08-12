package sql2http

import "testing"

var placeholderTests = []struct{
	ph   placeholderType
	in   string
	i    int
	want string
}{
	{placeholderSIMPLE|'?', ":name", 0, "?"},
	{placeholderNAMED|':', ":name", 0, ":name"},
	{placeholderNUMBER|':', ":name", 0, ":1"},
	{placeholderSIMPLE|'@', ":name", 1, "@"},
	{placeholderNAMED|'@', ":name", 1, "@name"},
	{placeholderNUMBER|'@', ":name", 1, "@2"},
}

func TestPlaceholderTranslate(t *testing.T) {
	for _, tc := range placeholderTests {
		got := tc.ph.translate(tc.in, tc.i)
		if got != tc.want {
			t.Errorf("type: %v %q -> %q; want %q", tc.ph, tc.in, got, tc.want)
		}
	}
}
