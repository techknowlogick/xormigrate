// +build postgresql

package xormigrate

import (
	_ "github.com/lib/pq"
)

func init() {
	databases = append(databases, database{
		name:    "postgres",
		connEnv: "PG_CONN_STRING",
	})
}
