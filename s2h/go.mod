module github.com/detaoin/sql2http/s2h

go 1.12

replace github.com/detaoin/sql2http/sql2http => ../sql2http

replace github.com/detaoin/sql2http/template => ../template

require (
	github.com/denisenkom/go-mssqldb v0.0.0-20190423183735-731ef375ac02
	github.com/detaoin/sql2http/sql2http v0.0.0
	github.com/detaoin/sql2http/template v0.0.0
	github.com/edsrzf/mmap-go v0.0.0-20170320065105-0bce6a688712 // indirect
	github.com/go-sql-driver/mysql v1.4.1
	github.com/lib/pq v1.1.0
	github.com/mattn/go-sqlite3 v1.10.0
	github.com/remyoudompheng/bigfft v0.0.0-20190321074620-2f0d2b0e0001 // indirect
	gopkg.in/yaml.v2 v2.2.2
	modernc.org/b v1.0.0 // indirect
	modernc.org/db v1.0.0 // indirect
	modernc.org/file v1.0.0 // indirect
	modernc.org/fileutil v1.0.0 // indirect
	modernc.org/golex v1.0.0 // indirect
	modernc.org/internal v1.0.0 // indirect
	modernc.org/lldb v1.0.0 // indirect
	modernc.org/mathutil v1.0.0 // indirect
	modernc.org/ql v1.0.0
	modernc.org/sortutil v1.0.0 // indirect
	modernc.org/strutil v1.0.0 // indirect
	modernc.org/zappy v1.0.0 // indirect
)
