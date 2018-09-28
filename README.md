sql2http: a simple SQL to HTTP gateway
======================================

**Warning**: this package is still in a highly volatile development stage,
use at your own risk.

A dead simple web interface for sequel databases.

Example
-------

First we compile/build the tool:

	go get -v github.com/detaoin/sql2http/cmd/s2h

For this example, you will need to compile it with cgo enabled, because
it uses the sqlite3 database driver which binds to the sqlite C library.

First we create a simple database:

	cat | sqlite3 test.db <<EOF
	BEGIN TRANSACTION;
	CREATE TABLE test ( num int, name text );
	INSERT INTO test VALUES(1,'foo');
	INSERT INTO test VALUES(2,'bar');
	INSERT INTO test VALUES(3,'baz');
	COMMIT;
	EOF

Then we create a simple config file:

	cat > s2h.yaml <<EOF
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

Or with the custom config format:

	cat > s2h.conf <<EOF
	sqlite3 test.db
	
	GET /
	names: SELECT * FROM test
	now:   SELECT datetime()
	
	GET /name/:id
	found: SELECT * FROM test WHERE num = :id
	EOF

Then we start the server (assuming it is in your PATH):

	s2h

You can now visit [localhost:8080/](http://localhost:8080/) or
[localhost:8080/name/2](http://localhost:8080/name/2).

cgo or no cgo?
--------------

The main package (`github.com/detaoin/sql2http`) let's you decide which
SQL drivers and templates you want to use, by importing them (maybe with
an _emtpy_ import) to have their `init` function register them.

However `cmd/s2h` imports both the SQL drivers and templates. The list
of drivers imported depends whether you are building with or without
cgo. The sqlite3 driver is imported only if compiling with cgo.

### List of SQL drivers (`cmd/s2h`)

- MySQL: [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)
- PostgreSQL: [github.com/lib/pq](https://github.com/lib/pq)
- SQLite3: [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)
