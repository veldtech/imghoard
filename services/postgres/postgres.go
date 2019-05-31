package imghoard

import (
	"log"
	"database/sql"
	migrations "github.com/mikibot/imghoard/services/postgres/migrations"
)

// Db is the database connection pool 
var Db *sql.DB

// InitDB creates the initial connection pool for the database.
func InitDB(connStr string) {
	db, err := sql.Open("postgres", connStr)
	if(err != nil) {
		log.Panicf("Unable to launch postgres with reason: %s", err)
	}
	Db = db
}

// RunMigrations runs migrations for the database.
func RunMigrations(connStr string, untilTime int64) error {
	if Db == nil {
		InitDB(connStr);
	}

	rows, err := Db.Query(`SELECT EXISTS (
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
		_, err = Db.Query(`CREATE TABLE public._databasemigrations (
			id BIGINT PRIMARY KEY);`)
		if err != nil {
			return err
		}
	}

	for _, migration := range []MigrationEntry{ 
		(*migrations.Initial)(nil), 
	} {
		// If migration exists, skip.
		rows, err := Db.Query(`SELECT 1 FROM public._databasemigrations WHERE id = $1`,
			migration.Id())
		if rows.Next() {
			continue;
		}

		if(migration.Id() >= untilTime) {
			err = migration.Up(Db)
			if err != nil {
				return err
			}

			_, err = Db.Query(`INSERT INTO public._databasemigrations (id) VALUES ($1);`,
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
	Up(*sql.DB) error
	Down(*sql.DB) error
}