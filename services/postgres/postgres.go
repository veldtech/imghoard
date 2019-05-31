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
func RunMigrations(connStr string, untilTime int64) {
	InitDB(connStr);

	for _, migration := range []MigrationEntry{ 
		(*migrations.Initial)(nil), 
	} {
		if(migration.Id() >= untilTime) {
			migration.Up(Db)
		}
	}
}

// MigrationEntry is used as a base interface for migrations
type MigrationEntry interface {
	Id() int64
	Up(*sql.DB) error
	Down(*sql.DB) error
}