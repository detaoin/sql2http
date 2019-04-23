module github.com/detaoin/sql2http/s2h

go 1.12

replace github.com/detaoin/sql2http/sql2http => ../sql2http

replace github.com/detaoin/sql2http/template => ../template

require (
	github.com/cznic/ql v1.2.0
	github.com/denisenkom/go-mssqldb v0.0.0-20190423183735-731ef375ac02
	github.com/detaoin/sql2http/sql2http v0.0.0
	github.com/detaoin/sql2http/template v0.0.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/lib/pq v1.1.0
	github.com/mattn/go-sqlite3 v1.10.0
	gopkg.in/yaml.v2 v2.2.2
)
