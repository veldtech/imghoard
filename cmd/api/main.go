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
	log.Println("Loading .env")
	err := godotenv.Load()
	if err != nil {
		log.Panicf("Error loading .env file: %s", err)
	}

	log.Print("Creating snowflake generator")
	snowflake.InitSnowflake()

	log.Println("Connecting to pg")

	connStr := createConnectionString()
	pg.InitDB(connStr)

	err = pg.Db.Ping()
	if err != nil {
		log.Panicf("Could not connect to postgres with connection string '%s': %s", connStr, err)
	}

	log.Println("Opening web service")

	portStr, exists := os.LookupEnv("PORT")
	port := 8080
	if exists {
		portInt, err := strconv.ParseInt(portStr, 0, 16)
		if err != nil {
			log.Panicf("Cannot create port from PORT environmental value %s", err)
		}
		port = int(portInt)
	}

	config := &atreugo.Config{
		Host: "0.0.0.0",
		Port: port,
	}

	server := atreugo.New(config)

	{
		baseURL := "127.0.0.1/"
		url, found := os.LookupEnv("URL_BASE")
		if found {
			baseURL = url
		}

		var imageView = images.ImageView{
			SpacesClient: spaces.New(),
			BaseURL:      baseURL,
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
