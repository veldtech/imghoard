package postgres

import (
	"log"
	"database/sql"
	migrations "github.com/mikibot/imghoard/services/postgres/migrations"
)

// InitDB creates the initial connection pool for the database.
func NewDB(connStr string) *sql.DB {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Panicf("Unable to launch postgres with reason: %s", err)
	}
	return db
}

// RunMigrations runs migrations for the database.
func RunMigrations(connStr string) {
	db := NewDB(connStr)

	for _, migration := range []MigrationEntry{ 
		(*migrations.Initial)(nil), 
	} {
		_ = migration.Up(db)
	}
}

// MigrationEntry is used as a base interface for migrations
type MigrationEntry interface {
	Down(*sql.DB) error

	Up(*sql.DB) error
}