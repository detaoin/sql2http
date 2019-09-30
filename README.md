sql2http: a simple SQL to HTTP gateway
======================================

**Warning**: this package is still in a highly volatile development stage,
use at your own risk.

A dead simple web interface for sequel databases.

Quick Start
-----------

First you will need some prerequisites:

- Go toolchain (available from [golang.org/dl/](https://golang.org/dl/))
- GCC or Clang C compiler (optional, however needed for this example with
  sqlite3)
- An editor (personally I use [vis](https://github.com/martanne/vis))

Then you can build the tool; assuming `go` and `gcc` are both in your
`PATH`, run in a shell (Bourne shell compatible, e.g. bash)

	$ go get -v git.sr.ht/~detaoin/sql2http/s2h

If everything went well, you should find the executable `s2h` (or
`s2h.exe` on Windows) inside directory `$GOPATH/bin/`.

Now we can create a sample database; here we assume you are using a
POSIX shell (bash for example):

	$ sqlite3 test.db <<EOF
	BEGIN TRANSACTION;
	CREATE TABLE test ( num int, name text );
	INSERT INTO test VALUES(1,'foo');
	INSERT INTO test VALUES(2,'bar');
	INSERT INTO test VALUES(3,'baz');
	COMMIT;
	EOF

Then we create a simple `s2h.yaml` config file:

	$ cat > s2h.yaml <<EOF
	db:
	  driver: sqlite3
	  options: test.db
	pages:
	  - pattern: /
	    method: GET
	    queries:
	      names: SELECT * FROM test
	      now:   SELECT datetime()
	  - pattern: /name/:id
	    method: GET
	    queries:
	      found: SELECT * FROM test WHERE num = :id
	EOF

Finally we start the server (assuming it is in your PATH):

	s2h

You can now visit [localhost:8080/](http://localhost:8080/) or
[localhost:8080/name/2](http://localhost:8080/name/2).

Configuration file format
-------------------------

The configuration file (either custom `.conf` or yaml format) is composed
of 2 main sections: the database connection parameters, and the configured
http pages with their respective SQL queries.

### Config: database specification

It is composed of a database driver selection
identifier, for example `sqlite3` for the
[github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)
go driver.

Then comes the specific driver options, most of the time consisting in
the database _address_ and user credentials.

For yaml configurations, the driver selection is given under key
`db/driver`, and the options under `db/options`.

For the custom configuration format, both are given on the first line
of the file; the driver selection being the first space separated word,
and the options the rest of the line.

For documentation on the specific driver options see the following links:

- sqlite3: [https://github.com/mattn/go-sqlite3#connection-string](https://github.com/mattn/go-sqlite3#connection-string)
- mysql: [https://github.com/go-sql-driver/mysql#dsn-data-source-name](https://github.com/go-sql-driver/mysql#dsn-data-source-name)
- postgres: [https://godoc.org/github.com/lib/pq#hdr-Connection_String_Parameters](https://godoc.org/github.com/lib/pq#hdr-Connection_String_Parameters)
- ql: [https://godoc.org/github.com/cznic/ql/driver](https://godoc.org/github.com/cznic/ql/driver)
- sqlserver: [https://github.com/denisenkom/go-mssqldb#connection-parameters-and-dsn](https://github.com/denisenkom/go-mssqldb#connection-parameters-and-dsn)

### Config: pages

Each page specification contains the following parameters:

- URL pattern
- HTTP method
- The list of SQL queries

The server will register a HTTP handler for each page specification; each
request matching the given URL pattern and HTTP method will trigger the
execution of all the specified SQL queries in a single transaction. If
it was successful the result structure (see below) is formatted using
the template matching the file extension of the requested URL.

The result structure is:

	type Result struct {
		Pattern string
		Params  map[string]interface{}
		Queries []Query
		Tables  Tables
		Request Request
		Time    time.Time // when the request was made
		Version string    // this package's version
	}

	type Request struct {
		URL    *url.URL
		Method string
		Header http.Header
	}

	type Query struct {
		Name   string
		Q      string
		Params []string
	}

	type Table struct {
		Name   string
		Header []string
		Rows   []Row
	}

	type Row struct {
		Header []string
		Values []interface{}
	}

The following default templates are compiled in `cmd/s2h`:

- `.html`: [git.sr.ht/~detaoin/sql2http/template/html](git.sr.ht/~detaoin/sql2http/template/html)
- `.tex`: [git.sr.ht/~detaoin/sql2http/template/tex](git.sr.ht/~detaoin/sql2http/template/tex)
- `.json`: [git.sr.ht/~detaoin/sql2http/template/json](git.sr.ht/~detaoin/sql2http/template/json)
- `.csv`: [git.sr.ht/~detaoin/sql2http/template/csv](git.sr.ht/~detaoin/sql2http/template/csv)
- `.tsv`: [git.sr.ht/~detaoin/sql2http/template/tsv](git.sr.ht/~detaoin/sql2http/template/tsv)
- `.xlsx`: [git.sr.ht/~detaoin/sql2http/template/xlsx](git.sr.ht/~detaoin/sql2http/template/xlsx)

If the requested URL has no file extension, it defaults to using the `.html` template.

### Config: SQL query parameters

In the SQL queries (of the configuration file) can use parameters
(e.g. `:name`) to interpolate the query with request specific values.

For example, the following page specification in the config file (yaml
for this example):

	- pattern: /name/:id
	  method: GET
	  queries:
	    found: SELECT * FROM test WHERE num = :id
	    form: SELECT :s

If a GET request is done for path `/name/5.html?s=something`, then the
first query _found_ will use `5` in the `WHERE` clause, and the _form_
query will return `something`.

The colon parameters are taken from the URL pattern first, then if not
found there from the url-encoded form data and POST data (in case of
POST requests).

### Config: templates

By default, the template used to render the `Result` struct is chosen
using the request path extension as key, falling back to `.html` of no
extension is specified.

However, first if there exists a file under `s2h.template/` having
the same relative path (except for extension) as the URL pattern which
matches a request, then that specific template is used.

For example, if a request matches pattern `/name/:id`, then depending on
the file extension, one of these templates is used instead of the default:

- `s2h.template/name/:id.html` if the request ends in `.html` or no extension,
- `s2h.template/name/:id.tex` if the request ends in `.tex`

cgo or no cgo?
--------------

The main package (`git.sr.ht/~detaoin/sql2http`) let's you decide which
SQL drivers and templates you want to use, by importing them (maybe with
an _emtpy_ import) to have their `init` function register them.

However `cmd/s2h` imports both the SQL drivers and templates. The list
of drivers imported depends whether you are building with or without
cgo. The sqlite3 driver is imported only if compiling with cgo.

List of SQL drivers (`cmd/s2h`)
-------------------------------

- MySQL: [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)
- PostgreSQL: [github.com/lib/pq](https://github.com/lib/pq)
- SQLite3: [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)
- ql: [github.com/cznic/ql](https://github.com/cznic/ql)
- SQL Server: [https://github.com/denisenkom/go-mssqldb](https://github.com/denisenkom/go-mssqldb)
