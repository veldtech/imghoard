package postgres

import (
	"database/sql"
)

type Client struct {
	*sql.DB
}

// InitDB creates the initial connection pool for the database.
func New(connStr string) (*Client, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(50)

	return &Client{
		DB: db,
	}, nil
}