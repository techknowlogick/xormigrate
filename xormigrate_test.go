package xormigrate

import (
	"os"
	"testing"

	"github.com/go-xorm/xorm"
	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/assert"
)

var databases []database

type database struct {
	name    string
	connEnv string
}

var migrations = []*Migration{
	{
		ID:   "201608301400",
		Desc: "Add Person",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(&Person{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(&Person{})
		},
	},
	{
		ID: "201608301430",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(&Pet{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(&Pet{})
		},
	},
}

var extendedMigrations = append(migrations, &Migration{
	ID: "201807221927",
	Migrate: func(tx *xorm.Engine) error {
		return tx.Sync2(&Book{})
	},
	Rollback: func(tx *xorm.Engine) error {
		return tx.DropTables(&Book{})
	},
})

type Person struct {
	ID   int `xorm:"id"`
	Name string
}

type Pet struct {
	Name     string
	PersonID int `xorm:"person_id"`
}

type Book struct {
	Name     string
	PersonID int `xorm:"person_id"`
}

func TestMigration(t *testing.T) {
	forEachDatabase(t, func(db *xorm.Engine) {
		m := New(db, migrations)
		m.SetLogger(nil)

		err := m.Migrate()
		assert.NoError(t, err)
		has, err := db.IsTableExist(&Person{})
		assert.NoError(t, err)
		assert.True(t, has)
		has, err = db.IsTableExist(&Pet{})
		assert.NoError(t, err)
		assert.True(t, has)
		assert.Equal(t, int64(2), tableCount(t, db))

		err = m.RollbackLast()
		assert.NoError(t, err)
		has, err = db.IsTableExist(&Person{})
		assert.NoError(t, err)
		assert.True(t, has)
		has, err = db.Exist(&Pet{})
		assert.Error(t, err)
		assert.False(t, has)
		assert.Equal(t, int64(1), tableCount(t, db))

		err = m.RollbackLast()
		assert.NoError(t, err)
		has, err = db.IsTableExist(&Person{})
		assert.NoError(t, err)
		assert.False(t, has)
		has, err = db.IsTableExist(&Pet{})
		assert.NoError(t, err)
		assert.False(t, has)
		assert.Equal(t, int64(0), tableCount(t, db))
	})
}

func TestMigrateTo(t *testing.T) {
	forEachDatabase(t, func(db *xorm.Engine) {
		m := New(db, extendedMigrations)

		err := m.MigrateTo("201608301430")
		assert.NoError(t, err)
		has, _ := db.IsTableExist(&Person{})
		assert.True(t, has)
		has, _ = db.IsTableExist(&Pet{})
		assert.True(t, has)
		has, _ = db.IsTableExist(&Book{})
		assert.False(t, has)
		assert.Equal(t, int64(2), tableCount(t, db))
	})
}

func TestRollbackTo(t *testing.T) {
	forEachDatabase(t, func(db *xorm.Engine) {
		m := New(db, extendedMigrations)

		// First, apply all migrations.
		err := m.Migrate()
		assert.NoError(t, err)
		has, _ := db.IsTableExist(&Person{})
		assert.True(t, has)
		has, _ = db.IsTableExist(&Pet{})
		assert.True(t, has)
		has, _ = db.IsTableExist(&Book{})
		assert.True(t, has)
		assert.Equal(t, int64(3), tableCount(t, db))

		// Rollback to the first migration: only the last 2 migrations are expected to be rolled back.
		err = m.RollbackTo("201608301400")
		assert.NoError(t, err)
		has, _ = db.IsTableExist(&Person{})
		assert.True(t, has)
		has, _ = db.IsTableExist(&Pet{})
		assert.False(t, has)
		has, _ = db.IsTableExist(&Book{})
		assert.False(t, has)
		assert.Equal(t, int64(1), tableCount(t, db))
	})
}

// If initSchema is defined, but no migrations are provided,
// then initSchema is executed.
func TestInitSchemaNoMigrations(t *testing.T) {
	forEachDatabase(t, func(db *xorm.Engine) {
		m := New(db, []*Migration{})
		m.InitSchema(func(tx *xorm.Engine) error {
			if err := tx.Sync2(&Person{}); err != nil {
				return err
			}
			return tx.Sync2(&Pet{}) // return error or nil
		})

		assert.NoError(t, m.Migrate())
		// assert.True(t, db.HasTable(&Person{}))
		// assert.True(t, db.HasTable(&Pet{}))
		assert.Equal(t, int64(1), tableCount(t, db))
	})
}

// If initSchema is defined and migrations are provided,
// then initSchema is executed and the migration IDs are stored,
// even though the relevant migrations are not applied.
func TestInitSchemaWithMigrations(t *testing.T) {
	forEachDatabase(t, func(db *xorm.Engine) {
		m := New(db, migrations)
		m.InitSchema(func(tx *xorm.Engine) error {
			return tx.Sync2(&Person{}) // return error or nil
		})

		assert.NoError(t, m.Migrate())
		has, _ := db.IsTableExist(&Person{})
		assert.True(t, has)
		has, _ = db.IsTableExist(&Pet{})
		assert.False(t, has)
		assert.Equal(t, int64(3), tableCount(t, db))
	})
}

// If the schema has already been initialised,
// then initSchema() is not executed, even if defined.
func TestInitSchemaAlreadyInitialised(t *testing.T) {
	type Car struct {
		ID int
	}

	forEachDatabase(t, func(db *xorm.Engine) {
		m := New(db, []*Migration{})

		// Migrate with empty initialisation
		m.InitSchema(func(tx *xorm.Engine) error {
			return nil
		})
		assert.NoError(t, m.Migrate())

		// Then migrate again, this time with a non empty initialisation
		// This second initialisation should not happen!
		m.InitSchema(func(tx *xorm.Engine) error {
			return tx.Sync2(&Car{}) // return error or nil
		})
		assert.NoError(t, m.Migrate())

		has, _ := db.IsTableExist(&Car{})
		assert.False(t, has)
		assert.Equal(t, int64(1), tableCount(t, db))
	})
}

// If the schema has not already been initialised,
// but any other migration has already been applied,
// then initSchema() is not executed, even if defined.
func TestInitSchemaExistingMigrations(t *testing.T) {
	type Car struct {
		ID int
	}

	forEachDatabase(t, func(db *xorm.Engine) {
		m := New(db, migrations)

		// Migrate without initialisation
		assert.NoError(t, m.Migrate())

		// Then migrate again, this time with a non empty initialisation
		// This initialisation should not happen!
		m.InitSchema(func(tx *xorm.Engine) error {
			return tx.Sync2(&Car{}) // return error or nil
		})
		assert.NoError(t, m.Migrate())

		has, _ := db.IsTableExist(&Car{})
		assert.False(t, has)
		assert.Equal(t, int64(2), tableCount(t, db))
	})
}

func TestMigrationIDDoesNotExist(t *testing.T) {
	forEachDatabase(t, func(db *xorm.Engine) {
		m := New(db, migrations)
		assert.Equal(t, ErrMigrationIDDoesNotExist, m.MigrateTo("1234"))
		assert.Equal(t, ErrMigrationIDDoesNotExist, m.RollbackTo("1234"))
		assert.Equal(t, ErrMigrationIDDoesNotExist, m.MigrateTo(""))
		assert.Equal(t, ErrMigrationIDDoesNotExist, m.RollbackTo(""))
	})
}

func TestMissingID(t *testing.T) {
	forEachDatabase(t, func(db *xorm.Engine) {
		migrationsMissingID := []*Migration{
			{
				Migrate: func(tx *xorm.Engine) error {
					return nil
				},
			},
		}

		m := New(db, migrationsMissingID)
		assert.Equal(t, ErrMissingID, m.Migrate())
	})
}

func TestReservedID(t *testing.T) {
	forEachDatabase(t, func(db *xorm.Engine) {
		migrationsReservedID := []*Migration{
			{
				ID: "SCHEMA_INIT",
				Migrate: func(tx *xorm.Engine) error {
					return nil
				},
			},
		}

		m := New(db, migrationsReservedID)
		_, isReservedIDError := m.Migrate().(*ReservedIDError)
		assert.True(t, isReservedIDError)
	})
}

func TestDuplicatedID(t *testing.T) {
	forEachDatabase(t, func(db *xorm.Engine) {
		migrationsDuplicatedID := []*Migration{
			{
				ID: "201705061500",
				Migrate: func(tx *xorm.Engine) error {
					return nil
				},
			},
			{
				ID: "201705061500",
				Migrate: func(tx *xorm.Engine) error {
					return nil
				},
			},
		}

		m := New(db, migrationsDuplicatedID)
		_, isDuplicatedIDError := m.Migrate().(*DuplicatedIDError)
		assert.True(t, isDuplicatedIDError)
	})
}

func TestEmptyMigrationList(t *testing.T) {
	forEachDatabase(t, func(db *xorm.Engine) {
		t.Run("with empty list", func(t *testing.T) {
			m := New(db, []*Migration{})
			err := m.Migrate()
			assert.Equal(t, ErrNoMigrationDefined, err)
		})

		t.Run("with nil list", func(t *testing.T) {
			m := New(db, nil)
			err := m.Migrate()
			assert.Equal(t, ErrNoMigrationDefined, err)
		})
	})
}

func tableCount(t *testing.T, db *xorm.Engine) (count int64) {
	count, err := db.Count(&Migration{})
	assert.NoError(t, err)
	return
}

func forEachDatabase(t *testing.T, fn func(database *xorm.Engine)) {
	if len(databases) == 0 {
		panic("No database chosen for testing!")
	}

	for _, database := range databases {
		db, err := xorm.NewEngine(database.name, os.Getenv(database.connEnv))
		assert.NoError(t, err, "Could not connect to database %s, %v", database.name, err)

		defer db.Close()

		// ensure tables do not exists
		assert.NoError(t, db.DropTables(&Migration{}, &Person{}, &Pet{}, &Book{}))

		fn(db)
	}
}
