package main

import (
	"fmt"
	_ "github.com/lib/pq"
	"github.com/mikibot/imghoard/config"
	pg "github.com/mikibot/imghoard/services/postgres"
	"github.com/mikibot/imghoard/services/snowflake"
	spaces "github.com/mikibot/imghoard/services/spaces"
	images "github.com/mikibot/imghoard/views"
	"github.com/savsgio/atreugo/v7"
	"log"
)

func main() {
	log.Println("Loading config")
	fileConfig, err := config.LoadFromFile("appconfig/secrets.json")
	if err != nil {
		log.Panicf("Error loading .env file: %s", err)
	}

	log.Print("Creating snowflake generator")
	idGenerator := snowflake.InitSnowflake()

	log.Println("Connecting to pg")
	connStr := createConnectionString(fileConfig)
	db := pg.NewDB(connStr)

	err = db.Ping()
	if err != nil {
		log.Panicf("Could not connect to postgres with connection string '%s': %s", connStr, err)
	}

	log.Println("Opening web service")

	httpConfig := &atreugo.Config{
		Host: "0.0.0.0",
		Port: 8080,
		Fasthttp: &atreugo.FasthttpConfig{
			MaxRequestBodySize: 20 * 1024 * 1024,
		},
	}

	server := atreugo.New(httpConfig)

	{
		var imageView = images.ImageView{
			BaseURL:      fileConfig.BaseURL,
			Db: 		  db,
			Generator:    idGenerator,
			SpacesClient: spaces.New(fileConfig, idGenerator),
		}

		server.Path("GET", "/images", imageView.GetImage)
		//server.Path("GET", "images/:id", imageView.GetImageByID)
		server.Path("POST", "/images", imageView.PostImage)
		server.Path("GET", "/tags/:id", imageView.GetTag)
	}

	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func createConnectionString(config config.Config) string {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s/%s",
		config.DatabaseUser,
		config.DatabasePass,
		config.DatabaseHost,
		config.DatabaseSchema)

	if !config.DatabaseUseSSL {
		connString += "?sslmode=disable"
	}

	return connString
}
