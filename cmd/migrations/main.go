package main

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	pg "github.com/mikibot/imghoard/services/postgres"
	snowflake "github.com/mikibot/imghoard/services/snowflake"
	spaces "github.com/mikibot/imghoard/services/spaces"
	images "github.com/mikibot/imghoard/views"
	"github.com/savsgio/atreugo/v7"
)

func main() {
	log.Print("starting migrations")
	// TODO: implement migrations
}

func createConnectionString() string {
	connStr := "postgres://"
	user, exists := os.LookupEnv("PG_USER")
	if exists {
		connStr += user
	}

	pass, exists := os.LookupEnv("PG_PASS")
	if exists {
		connStr += ":" + pass
	}

	host, exists := os.LookupEnv("PG_HOST")
	if exists {
		connStr += "@" + host
	}

	db, exists := os.LookupEnv("PG_DB")
	if exists {
		connStr += "/" + db
	}

	ssl, exists := os.LookupEnv("PG_SSLMODE")
	if exists {
		connStr += "?sslmode=" + ssl
	}
	return connStr
}
