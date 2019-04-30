package sql2http

import "testing"

var testcases = []struct {
	input string
	toks  []item
}{
	{
		"",
		[]item{},
	},
	{
		"SELECT /* comment /* nested */ ... */ \"me\" FROM them \t-- comment\n'a string' + 0.12e-5+.012",
		[]item{
			item{typ: itemIdentifier, val: "SELECT"},
			item{typ: itemSpace, val: " "},
			item{typ: itemBlockComment, val: "/* comment /* nested */ ... */"},
			item{typ: itemSpace, val: " "},
			item{typ: itemQuotedIdentifier, val: "\"me\""},
			item{typ: itemSpace, val: " "},
			item{typ: itemIdentifier, val: "FROM"},
			item{typ: itemSpace, val: " "},
			item{typ: itemIdentifier, val: "them"},
			item{typ: itemSpace, val: " \t"},
			item{typ: itemComment, val: "-- comment\n"},
			item{typ: itemStringLiteral, val: "'a string'"},
			item{typ: itemSpace, val: " "},
			item{typ: itemOperator, val: "+"},
			item{typ: itemSpace, val: " "},
			item{typ: itemNumeric, val: "0.12e-5"},
			item{typ: itemOperator, val: "+"},
			item{typ: itemNumeric, val: ".012"},
		},
	},
}

func TestLexSQL(t *testing.T) {
	for _, tc := range testcases {
		l := lexSQL(tc.input)
		for _, want := range tc.toks {
			got, ok := <-l.items
			if !ok {
				t.Fatalf("expecting %v; end of stream", want)
			}
			if got.typ != want.typ || got.val != want.val {
				t.Fatalf("expecting %v; got %v", want, got)
			}
		}
		got, ok := <-l.items
		if !ok || got.typ != itemEOF {
			t.Fatalf("expecting EOF")
		}
		got, ok = <-l.items
		if ok {
			t.Fatalf("superfluous token: %v", got)
		}
	}
}
