module src.techknowlogick.com/xormigrate

go 1.12

require (
	cloud.google.com/go v0.37.2 // indirect
	github.com/denisenkom/go-mssqldb v0.0.0-20190423194141-731ef375ac02
	github.com/go-sql-driver/mysql v1.4.1
	github.com/go-xorm/builder v0.3.4 // indirect
	github.com/go-xorm/xorm v0.7.1
	github.com/joho/godotenv v1.3.0
	github.com/lib/pq v1.1.0
	github.com/mattn/go-sqlite3 v1.10.0
	github.com/stretchr/testify v1.3.0
	golang.org/x/crypto v0.0.0-20190325154230-a5d413f7728c // indirect
	google.golang.org/appengine v1.5.0 // indirect
)

replace github.com/go-xorm/xorm => github.com/go-xorm/xorm v0.7.2-0.20190330194841-617e0ae295d7fd8a4ea48f6c782781f6cc367c7e

replace github.com/golang/lint => golang.org/x/lint v0.0.0-20190330180304-d0100b6
