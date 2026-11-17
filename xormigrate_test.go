package xormigrate

import (
	"context"
	"errors"
	"os"
	"testing"

	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/assert"
	"xorm.io/xorm"
)

var databases []database

type database struct {
	name    string
	connEnv string
}

var migrations = []*Migration{
	{
		ID:          "201608301400",
		Description: "Add Person",
		Migrate: func(_ context.Context, tx *xorm.Engine) error {
			return tx.Sync(&Person{})
		},
		Rollback: func(_ context.Context, tx *xorm.Engine) error {
			return tx.DropTables(&Person{})
		},
	},
	{
		ID: "201608301430",
		Migrate: func(_ context.Context, tx *xorm.Engine) error {
			return tx.Sync(&Pet{})
		},
		Rollback: func(_ context.Context, tx *xorm.Engine) error {
			return tx.DropTables(&Pet{})
		},
	},
}

var migrationsSession = []*Migration{
	{
		ID:          "201608301400",
		Description: "Add Person",
		MigrateSession: func(_ context.Context, tx *xorm.Session) error {
			return tx.Sync(&Person{})
		},
		RollbackSession: func(_ context.Context, tx *xorm.Session) error {
			return tx.DropTable(&Person{})
		},
	},
	{
		ID: "201608301430",
		MigrateSession: func(_ context.Context, tx *xorm.Session) error {
			return tx.Sync(&Pet{})
		},
		RollbackSession: func(_ context.Context, tx *xorm.Session) error {
			return tx.DropTable(&Pet{})
		},
	},
}

var extendedMigrations = append(migrations, &Migration{
	ID: "201807221927",
	Migrate: func(_ context.Context, tx *xorm.Engine) error {
		return tx.Sync(&Book{})
	},
	Rollback: func(_ context.Context, tx *xorm.Engine) error {
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

		err := m.Migrate(t.Context())
		assert.NoError(t, err)
		has, err := db.IsTableExist(&Person{})
		assert.NoError(t, err)
		assert.True(t, has)
		has, err = db.IsTableExist(&Pet{})
		assert.NoError(t, err)
		assert.True(t, has)
		assert.Equal(t, int64(2), tableCount(t, db))

		err = m.RollbackLast(t.Context())
		assert.NoError(t, err)
		has, err = db.IsTableExist(&Person{})
		assert.NoError(t, err)
		assert.True(t, has)
		has, err = db.Exist(&Pet{})
		assert.Error(t, err)
		assert.False(t, has)
		assert.Equal(t, int64(1), tableCount(t, db))

		err = m.RollbackLast(t.Context())
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

		err := m.MigrateTo(t.Context(), "201608301430")
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
		err := m.Migrate(t.Context())
		assert.NoError(t, err)
		has, _ := db.IsTableExist(&Person{})
		assert.True(t, has)
		has, _ = db.IsTableExist(&Pet{})
		assert.True(t, has)
		has, _ = db.IsTableExist(&Book{})
		assert.True(t, has)
		assert.Equal(t, int64(3), tableCount(t, db))

		// Rollback to the first migration: only the last 2 migrations are expected to be rolled back.
		err = m.RollbackTo(t.Context(), "201608301400")
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
		m.InitSchema(func(_ context.Context, tx *xorm.Engine) error {
			if err := tx.Sync(&Person{}); err != nil {
				return err
			}
			return tx.Sync(&Pet{}) // return error or nil
		})

		assert.NoError(t, m.Migrate(t.Context()))
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
		m.InitSchema(func(_ context.Context, tx *xorm.Engine) error {
			return tx.Sync(&Person{}) // return error or nil
		})

		assert.NoError(t, m.Migrate(t.Context()))
		has, _ := db.IsTableExist(&Person{})
		assert.True(t, has)
		has, _ = db.IsTableExist(&Pet{})
		assert.False(t, has)
		assert.Equal(t, int64(3), tableCount(t, db))
	})
}

// If the schema has already been initialized,
// then initSchema() is not executed, even if defined.
func TestInitSchemaAlreadyInitialised(t *testing.T) {
	type Car struct {
		ID int
	}

	forEachDatabase(t, func(db *xorm.Engine) {
		m := New(db, []*Migration{})

		// Migrate with empty initialisation
		m.InitSchema(func(_ context.Context, tx *xorm.Engine) error {
			return nil
		})
		assert.NoError(t, m.Migrate(t.Context()))

		// Then migrate again, this time with a non empty initialisation
		// This second initialisation should not happen!
		m.InitSchema(func(_ context.Context, tx *xorm.Engine) error {
			return tx.Sync(&Car{}) // return error or nil
		})
		assert.NoError(t, m.Migrate(t.Context()))

		has, _ := db.IsTableExist(&Car{})
		assert.False(t, has)
		assert.Equal(t, int64(1), tableCount(t, db))
	})
}

// If the schema has not already been initialized,
// but any other migration has already been applied,
// then initSchema() is not executed, even if defined.
func TestInitSchemaExistingMigrations(t *testing.T) {
	type Car struct {
		ID int
	}

	forEachDatabase(t, func(db *xorm.Engine) {
		m := New(db, migrations)

		// Migrate without initialisation
		assert.NoError(t, m.Migrate(t.Context()))

		// Then migrate again, this time with a non empty initialisation
		// This initialisation should not happen!
		m.InitSchema(func(_ context.Context, tx *xorm.Engine) error {
			return tx.Sync(&Car{}) // return error or nil
		})
		assert.NoError(t, m.Migrate(t.Context()))

		has, _ := db.IsTableExist(&Car{})
		assert.False(t, has)
		assert.Equal(t, int64(2), tableCount(t, db))
	})
}

func TestInitSchemaError(t *testing.T) {
	errInit := errors.New("init schema failed")
	forEachDatabase(t, func(db *xorm.Engine) {
		m := New(db, migrations)
		m.InitSchema(func(context.Context, *xorm.Engine) error { return errInit })

		// a failed init must not mark any migration as done
		assert.ErrorIs(t, m.Migrate(t.Context()), errInit)
		assert.Equal(t, int64(0), tableCount(t, db))
	})
}

func TestMigrationIDDoesNotExist(t *testing.T) {
	forEachDatabase(t, func(db *xorm.Engine) {
		m := New(db, migrations)
		assert.Equal(t, ErrMigrationIDDoesNotExist, m.MigrateTo(t.Context(), "1234"))
		assert.Equal(t, ErrMigrationIDDoesNotExist, m.RollbackTo(t.Context(), "1234"))
		assert.Equal(t, ErrMigrationIDDoesNotExist, m.MigrateTo(t.Context(), ""))
		assert.Equal(t, ErrMigrationIDDoesNotExist, m.RollbackTo(t.Context(), ""))
	})
}

func TestMissingID(t *testing.T) {
	forEachDatabase(t, func(db *xorm.Engine) {
		migrationsMissingID := []*Migration{
			{
				Migrate: func(_ context.Context, tx *xorm.Engine) error {
					return nil
				},
			},
		}

		m := New(db, migrationsMissingID)
		assert.Equal(t, ErrMissingID, m.Migrate(t.Context()))
	})
}

func TestMissingMigrate(t *testing.T) {
	forEachDatabase(t, func(db *xorm.Engine) {
		m := New(db, []*Migration{{ID: "201608301400"}})
		assert.ErrorIs(t, m.Migrate(t.Context()), ErrMissingMigrate)
	})
}

func TestMigrationTableError(t *testing.T) {
	forEachDatabase(t, func(db *xorm.Engine) {
		m := New(db, migrations)
		assert.NoError(t, m.Migrate(t.Context()))

		// a failing query must not look like a migration that did not run
		assert.NoError(t, db.DropTables(&Migration{}))
		err := m.RollbackLast(t.Context())
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrNoRunMigration)
		assert.Error(t, m.RollbackTo(t.Context(), "201608301400"))
	})
}

func TestMigrationErrorIsWrapped(t *testing.T) {
	errMigrate := errors.New("migrate failed")
	forEachDatabase(t, func(db *xorm.Engine) {
		m := New(db, []*Migration{{
			ID:      "201608301400",
			Migrate: func(context.Context, *xorm.Engine) error { return errMigrate },
		}})
		assert.ErrorIs(t, m.Migrate(t.Context()), errMigrate)

		m = New(db, []*Migration{{
			ID:             "201608301400",
			MigrateSession: func(context.Context, *xorm.Session) error { return errMigrate },
		}})
		assert.ErrorIs(t, m.Migrate(t.Context()), errMigrate)
	})
}

func TestMigrateStopsWhenContextIsCanceled(t *testing.T) {
	forEachDatabase(t, func(db *xorm.Engine) {
		ctx, cancel := context.WithCancel(t.Context())
		var ran []string
		m := New(db, []*Migration{
			{ID: "1", Migrate: func(context.Context, *xorm.Engine) error { ran = append(ran, "1"); cancel(); return nil }},
			{ID: "2", Migrate: func(context.Context, *xorm.Engine) error { ran = append(ran, "2"); return nil }},
		})
		assert.ErrorIs(t, m.Migrate(ctx), context.Canceled)
		assert.Equal(t, []string{"1"}, ran)

		// the finished migration is recorded, so the next run goes on after it
		assert.NoError(t, m.Migrate(t.Context()))
		assert.Equal(t, []string{"1", "2"}, ran)
	})
}

func TestMigrationGetsContext(t *testing.T) {
	forEachDatabase(t, func(db *xorm.Engine) {
		ctx, cancel := context.WithCancel(t.Context())
		m := New(db, []*Migration{{
			ID: "1",
			Migrate: func(ctx context.Context, _ *xorm.Engine) error {
				cancel()
				return ctx.Err()
			},
		}})
		assert.ErrorIs(t, m.Migrate(ctx), context.Canceled)
		assert.Equal(t, int64(0), tableCount(t, db))

		// a session migration gets a session that uses ctx
		ctx, cancel = context.WithCancel(t.Context())
		m = New(db, []*Migration{{
			ID: "1",
			MigrateSession: func(_ context.Context, sess *xorm.Session) error {
				cancel()
				_, err := sess.Count(&Migration{})
				return err
			},
		}})
		assert.ErrorIs(t, m.Migrate(ctx), context.Canceled)
		assert.Equal(t, int64(0), tableCount(t, db))
	})
}

func TestRollbackStopsWhenContextIsCanceled(t *testing.T) {
	forEachDatabase(t, func(db *xorm.Engine) {
		ctx, cancel := context.WithCancel(t.Context())
		noop := func(context.Context, *xorm.Engine) error { return nil }
		m := New(db, []*Migration{
			{ID: "1", Migrate: noop, Rollback: noop},
			{ID: "2", Migrate: noop, Rollback: noop},
			{ID: "3", Migrate: noop, Rollback: func(context.Context, *xorm.Engine) error { cancel(); return nil }},
		})
		assert.NoError(t, m.Migrate(t.Context()))

		// 3 is rolled back and its record is gone, 2 is left alone
		assert.ErrorIs(t, m.RollbackTo(ctx, "1"), context.Canceled)
		assert.Equal(t, int64(2), tableCount(t, db))
	})
}

func TestReservedID(t *testing.T) {
	forEachDatabase(t, func(db *xorm.Engine) {
		migrationsReservedID := []*Migration{
			{
				ID: "SCHEMA_INIT",
				Migrate: func(_ context.Context, tx *xorm.Engine) error {
					return nil
				},
			},
		}

		m := New(db, migrationsReservedID)
		assert.ErrorAs(t, m.Migrate(t.Context()), new(*ReservedIDError))
	})
}

func TestDuplicatedID(t *testing.T) {
	forEachDatabase(t, func(db *xorm.Engine) {
		migrationsDuplicatedID := []*Migration{
			{
				ID: "201705061500",
				Migrate: func(_ context.Context, tx *xorm.Engine) error {
					return nil
				},
			},
			{
				ID: "201705061500",
				Migrate: func(_ context.Context, tx *xorm.Engine) error {
					return nil
				},
			},
		}

		m := New(db, migrationsDuplicatedID)
		assert.ErrorAs(t, m.Migrate(t.Context()), new(*DuplicatedIDError))
	})
}

func TestEmptyMigrationList(t *testing.T) {
	forEachDatabase(t, func(db *xorm.Engine) {
		t.Run("with empty list", func(t *testing.T) {
			m := New(db, []*Migration{})
			err := m.Migrate(t.Context())
			assert.Equal(t, ErrNoMigrationDefined, err)
		})

		t.Run("with nil list", func(t *testing.T) {
			m := New(db, nil)
			err := m.Migrate(t.Context())
			assert.Equal(t, ErrNoMigrationDefined, err)
		})
	})
}

func TestAllowLong(t *testing.T) {
	forEachDatabase(t, func(db *xorm.Engine) {
		t.Run("without AllowLong", func(t *testing.T) {
			m := New(db, []*Migration{{
				ID:   "201608301430",
				Long: true,
				Migrate: func(_ context.Context, tx *xorm.Engine) error {
					return tx.Sync(&Pet{})
				},
				Rollback: func(_ context.Context, tx *xorm.Engine) error {
					return tx.DropTables(&Pet{})
				},
			}})
			err := m.Migrate(t.Context())
			assert.Nil(t, err)
			assert.Equal(t, int64(0), tableCount(t, db))
		})

		t.Run("with AllowLong", func(t *testing.T) {
			m := New(db, []*Migration{{
				ID:   "201608301430",
				Long: true,
				Migrate: func(_ context.Context, tx *xorm.Engine) error {
					return tx.Sync(&Pet{})
				},
				Rollback: func(_ context.Context, tx *xorm.Engine) error {
					return tx.DropTables(&Pet{})
				},
			}})
			m.AllowLong(true)
			err := m.Migrate(t.Context())
			assert.Nil(t, err)
			assert.Equal(t, int64(1), tableCount(t, db))
		})
	})
}

func TestMigrationSession(t *testing.T) {
	forEachDatabase(t, func(db *xorm.Engine) {
		m := New(db, migrationsSession)

		err := m.Migrate(t.Context())
		assert.NoError(t, err)
		has, err := db.IsTableExist(&Person{})
		assert.NoError(t, err)
		assert.True(t, has)
		has, err = db.IsTableExist(&Pet{})
		assert.NoError(t, err)
		assert.True(t, has)
		assert.Equal(t, int64(2), tableCount(t, db))

		err = m.RollbackLast(t.Context())
		assert.NoError(t, err)
		has, err = db.IsTableExist(&Person{})
		assert.NoError(t, err)
		assert.True(t, has)
		has, err = db.Exist(&Pet{})
		assert.Error(t, err)
		assert.False(t, has)
		assert.Equal(t, int64(1), tableCount(t, db))

		err = m.RollbackLast(t.Context())
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

func tableCount(t *testing.T, db *xorm.Engine) (count int64) {
	count, err := db.Count(&Migration{})
	assert.NoError(t, err)
	return count
}

func forEachDatabase(t *testing.T, fn func(database *xorm.Engine)) {
	if len(databases) == 0 {
		t.Fatal("No database chosen for testing!")
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
