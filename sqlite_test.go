// +build sqlite

package xormigrate

import (
	_ "github.com/mattn/go-sqlite3"
)

func init() {
	databases = append(databases, database{
		name:    "sqlite3",
		connEnv: "SQLITE_CONN_STRING",
	})
}
