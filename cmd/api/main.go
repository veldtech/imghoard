package main

import (
	framework "github.com/mikibot/imghoard/framework"
	imagehandler "github.com/mikibot/imghoard/services/imagehandler"
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
	loadEnv()

	log.Print("Creating snowflake generator")
	uuidGen := snowflake.InitSnowflake()

	log.Println("Connecting to pg")
	connStr := createConnectionString()
	db, err := pg.New(connStr)
	if err != nil {
		log.Panicf("Could not connect to PostgreSQL because of reason: %s", err)
	}

	spacesAccess, exists := os.LookupEnv("DO_ACCESS")
	if !exists {
		log.Panic("Could not load env variable DO_ACCESS")
		os.Exit(1)
	}
	spacesSecret, exists := os.LookupEnv("DO_SECRET")
	if !exists {
		log.Panic("Could not load env variable DO_SECRET")
		os.Exit(1)
	}
	spacesEndpoint, exists := os.LookupEnv("DO_ENDPOINT")
	if !exists {
		log.Panic("Could not load env variable DO_ENDPOINT")
		os.Exit(1)
	}
	spacesFolder, exists := os.LookupEnv("DO_FOLDER")
	if !exists {
		log.Panic("Could not load env variable DO_FOLDER")
		os.Exit(1)
	}
	spacesBucket, exists := os.LookupEnv("DO_BUCKET")
	if !exists {
		log.Panic("Could not load env variable DO_BUCKET")
		os.Exit(1)
	}

	spacesClient := spaces.New(spaces.Config{
		AccessKey: spacesAccess,
		SecretKey: spacesSecret,
		Endpoint: spacesEndpoint,
		Folder: spacesFolder,
		Bucket: spacesBucket,
	}, uuidGen)

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

	server := atreugo.New(&atreugo.Config{
		Host: "0.0.0.0",
		Port: port,
	})

	{
		baseURL := "127.0.0.1/"
		url, found := os.LookupEnv("URL_BASE")
		if found {
			baseURL = url
		}

		var imageView = images.ImageView{
			BaseURL: baseURL,
			Handler: imagehandler.New(baseURL, spacesClient, db),
		}

		var mockImageView = images.ImageView {
			BaseURL: baseURL,
			Handler: imagehandler.NewMock(baseURL, spacesClient, db),
		}

		{ // GetImage Route
 			view := framework.New(imageView.GetImage)
			view.AddTenancy("testing", mockImageView.GetImage)
			server.Path("GET", "/images", view.Route)
		}

		{ // PostImage Route
			view := framework.New(imageView.PostImage)
			view.AddTenancy("testing", mockImageView.PostImage)
			server.Path("POST", "/images", view.Route)
		}

		{ // GetTag Route
			view := framework.New(imageView.GetTag)
			view.AddTenancy("testing", mockImageView.GetTag)
			server.Path("GET", "/tags/:id", view.Route)
		}
	}
	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func loadEnv() {
	log.Println("Loading .env")
	err := godotenv.Load()
	if err != nil {
		log.Panicf("Error loading .env file: %s", err)
		os.Exit(1)
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
