//go:build sqlserver
// +build sqlserver

package xormigrate

import (
	_ "github.com/denisenkom/go-mssqldb"
)

func init() {
	databases = append(databases, database{
		name:    "mssql",
		connEnv: "SQLSERVER_CONN_STRING",
	})
}
