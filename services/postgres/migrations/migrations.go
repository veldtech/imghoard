package imghoard

import (
	"github.com/mikibot/imghoard/services/postgres"
)

// RunMigrations runs migrations for the database.
func (db *postgres.Client) RunMigrations(connStr string, untilTime int64) error {
	rows, err := db.Query(`SELECT EXISTS (
		SELECT 1
		FROM   information_schema.tables 
		WHERE  table_schema = 'public'
		AND    table_name = '_databasemigrations'
		);`)
	if err != nil {
		return err
	}

	var exists bool
	if rows.Next() {
		rows.Scan(&exists)
	}
	if !exists {
		_, err = db.Query(`CREATE TABLE public._databasemigrations (
			id BIGINT PRIMARY KEY);`)
		if err != nil {
			return err
		}
	}

	for _, migration := range []MigrationEntry{
		(*migrations.Initial)(nil),
	} {
		// If migration exists, skip.
		rows, err := db.Query(`SELECT 1 FROM public._databasemigrations WHERE id = $1`,
			migration.Id())
		if rows.Next() {
			continue
		}

		if migration.Id() >= untilTime {
			err = migration.Up(db)
			if err != nil {
				return err
			}

			_, err = db.Query(`INSERT INTO public._databasemigrations (id) VALUES ($1);`,
				migration.Id())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// MigrationEntry is used as a base interface for migrations
type MigrationEntry interface {
	Id() int64
	Up(*PostgresClient) error
	Down(*PostgresClient) error
}
