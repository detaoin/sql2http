package sql2http

// Query represents a single query, with its name.
type Query struct {
	Name string
	Q    string
}

// Row represents a row of query result. The Headers are present for
// column name lookup of the row data.
type Row struct {
	Header []string
	Values []interface{}
}

// Get returns the value of the Row for the given column key. If it
// doesn't exist, nil is returned.
func (r Row) Get(key string) interface{} {
	for i, h := range r.Header {
		if key == h {
			return r.Values[i]
		}
	}
	return nil
}

type Table struct {
	Name   string
	Header []string
	Rows   []Row
}

type Tables []Table

// Get returns the Table with given name. If it doesn't exist, and empty
// Table is returned.
func (t Tables) Get(name string) Table {
	for _, table := range t {
		if table.Name == name {
			return table
		}
	}
	return Table{}
}
