# Xormigrate

## Supported databases

It supports any of the databases Xorm supports:

- PostgreSQL
- MySQL
- SQLite
- Microsoft SQL Server

## Installing

```bash
go get -u src.techknowlogick.com/xormigrate
```

## Usage

```go
package main

import (
	"log"

	"src.techknowlogick.com/xormigrate"

	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := xorm.NewEngine("sqlite3", "mydb.sqlite3")
	if err != nil {
		log.Fatal(err)
	}

	m := xormigrate.New(db, []*xormigrate.Migration{
		// create persons table
		{
			ID: "201608301400",
			Migrate: func(tx *xorm.Engine) error {
				// it's a good pratice to copy the struct inside the function,
				// so side effects are prevented if the original struct changes during the time
				type Person struct {
					Name string
				}
				return tx.Sync2(&Person{})
			},
			Rollback: func(tx *xorm.Engine) error {
				return tx.DropTables(&Person{})
			},
		},
		// add age column to persons
		{
			ID: "201608301415",
			Migrate: func(tx *xorm.Engine) error {
				// when table already exists, it just adds fields as columns
				type Person struct {
					Age int
				}
				return tx.Sync2(&Person{})
			},
			Rollback: func(tx *xorm.Engine) error {
				// Note: Column dropping in sqlite is not support, and you will need to do this manually
				_, err = sess.Exec("ALTER TABLE person DROP COLUMN age")
				if err != nil {
					return fmt.Errorf("Drop column failed: %v", err)
				}
				return nil
			},
		},
		// add pets table
		{
			ID: "201608301430",
			Migrate: func(tx *xorm.Engine) error {
				type Pet struct {
					Name     string
					PersonID int
				}
				return tx.Sync2(&Pet{})
			},
			Rollback: func(tx *xorm.Engine) error {
				return tx.DropTables(&Pet{})
			},
		},
	})

	if err = m.Migrate(); err != nil {
		log.Fatalf("Could not migrate: %v", err)
	}
	log.Printf("Migration did run successfully")
}
```

## Having a separated function for initializing the schema

If you have a lot of migrations, it can be a pain to run all them, as example,
when you are deploying a new instance of the app, in a clean database.
To prevent this, you can set a function that will run if no migration was run
before (in a new clean database). Remember to create everything here, all tables,
foreign keys and what more you need in your app.

```go
type Person struct {
	Name string
	Age int
}

type Pet struct {
	Name     string
	PersonID int
}

m := xormigrate.New(db, []*xormigrate.Migration{
    // you migrations here
})

m.InitSchema(func(tx *xorm.Engine) error {
	err := tx.sync2(
		&Person{},
		&Pet{},
		// all other tables of your app
	)
	if err != nil {
		return err
	}
	return nil
})
```

## Credits

* Based on [Gormigrate][gormmigrate]
* Uses [Xorm][xorm]

[xorm]: http://github.com/go-xorm/xorm/
[gormmigrate]: https://github.com/go-gormigrate/gormigrate
