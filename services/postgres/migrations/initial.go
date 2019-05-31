package imghoard

import (
	"database/sql"
)

// Initial is a migration.
type Initial struct {}

func (m Initial) Up(db *sql.DB) error {
	return nil
}

func (m Initial) Down(db *sql.DB) error {
	return nil
}