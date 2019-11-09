package imghoard

import (
	"fmt"
	"time"
	"github.com/mikibot/imghoard/services/postgres"
)

// Initial is a migration.
type Initial struct{}

func (m Initial) Id() int64 {
	return time.Date(2019, 5, 31, 23, 36, 0, 0, nil).Unix()
}

func (m Initial) Up(db *postgres.Client) error {
	_, err := db.Query(`SELECT 'CREATE DATABASE imghoard'
		WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'imghoard')`)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	_, err = db.Query(`CREATE TABLE user (
		id BIGINT PRIMARY KEY);`)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	_, err = db.Query(`CREATE TABLE image (
		id BIGINT PRIMARY KEY,
		contentType TEXT NOT NULL,
		author BIGINT REFERENCES user(id));`)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	_, err = db.Query(`CREATE TABLE tag (
		id BIGINT PRIMARY KEY,
		name TEXT);`)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	_, err = db.Query(`CREATE TABLE image_tags (
		image_id BIGINT REFERENCES image(id),
		tag_id BIGINT REFERENCES tag(id));`)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	return nil
}

func (m Initial) Down(db *postgres.Client) error {
	_, err := db.Query(`DROP TABLE image_tags;`)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	_, err = db.Query(`DROP TABLE tag;`)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	_, err = db.Query(`DROP TABLE image;`)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	_, err = db.Query(`DROP TABLE user;`)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	_, err = db.Query(`DROP DATABASE imghoard;`)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	return nil
}
