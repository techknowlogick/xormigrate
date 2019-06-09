module src.techknowlogick.com/xormigrate

go 1.12

require (
	github.com/denisenkom/go-mssqldb v0.0.0-20190515214422-eb9f6a1743f3
	github.com/go-sql-driver/mysql v1.4.1
	github.com/go-xorm/builder v0.3.4 // indirect
	github.com/go-xorm/xorm v0.7.1
	github.com/joho/godotenv v1.3.0
	github.com/lib/pq v1.1.1
	github.com/mattn/go-sqlite3 v1.10.0
	github.com/stretchr/testify v1.3.0
	google.golang.org/appengine v1.6.1 // indirect
)

replace github.com/go-xorm/xorm => github.com/go-xorm/xorm v0.7.2-0.20190330194841-617e0ae295d7fd8a4ea48f6c782781f6cc367c7e

replace github.com/golang/lint => golang.org/x/lint v0.0.0-20190330180304-d0100b6
