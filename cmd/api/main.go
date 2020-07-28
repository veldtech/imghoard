package main

import (
	"fmt"
	"github.com/json-iterator/go/extra"
	"github.com/mikibot/imghoard/middleware"
	models "github.com/mikibot/imghoard/models"
	"log"

	framework "github.com/mikibot/imghoard/framework"
	imagehandler "github.com/mikibot/imghoard/services/imagehandler"

	"github.com/getsentry/sentry-go"
	_ "github.com/lib/pq"
	"github.com/mikibot/imghoard/config"
	pg "github.com/mikibot/imghoard/services/postgres"
	"github.com/mikibot/imghoard/services/snowflake"
	spaces "github.com/mikibot/imghoard/services/spaces"
	images "github.com/mikibot/imghoard/views"
	"github.com/savsgio/atreugo/v11"
)

func corsMiddleware(ctx *atreugo.RequestCtx) error {
	ctx.Response.Header.Add("Access-Control-Allow-Origin", "*")
	return ctx.Next()
}

func main() {
	extra.SetNamingStrategy(extra.LowerCaseWithUnderscores)

	log.Println("Loading config")
	fileConfig, err := config.LoadFromFile("appconfig/secrets.json")
	if err != nil {
		log.Panicf("Error loading .env file: %s", err)
	}

	err = sentry.Init(sentry.ClientOptions{Dsn: fileConfig.SentryDsn})
	if err != nil {
		log.Println("Sentry was not initialized for reason: " + err.Error())
	}

	log.Print("Creating snowflake generator")
	idGenerator, err := snowflake.New()
	if err != nil {
		log.Panic(err)
	}

	log.Println("Connecting to pg")
	connStr := createConnectionString(fileConfig)
	db, err := pg.New(connStr)
	if err != nil {
		log.Panic(err)
	}

	err = db.Ping()
	if err != nil {
		log.Panic(err)
	}

	spacesClient := spaces.New(fileConfig, idGenerator)

	log.Println("Opening web service")

	addr := fmt.Sprintf("0.0.0.0:%d", 8080)
	server := atreugo.New(atreugo.Config{
		Addr: addr,
		MaxRequestBodySize: 20 * 2048 * 2048 * 2048,
		ErrorView: func(ctx *atreugo.RequestCtx, err error, i int) {

			_ = middleware.HandleSentryEvent(ctx, err)
			_ = ctx.JSONResponse(models.ErrorResponse{
				Message: "Internal Server Error",
				RequestId: ctx.ID(),
			})
		},
	})

	server.UseBefore(corsMiddleware)
	server.UseAfter(middleware.NewErrorMapper())

	{
		var imageView = images.ImageView{
			BaseUrl: fileConfig.BaseURL,
			Handler: imagehandler.New(fileConfig.BaseURL, spacesClient, db),
		}

		var mockImageView = images.ImageView{
			BaseUrl: fileConfig.BaseURL,
			Handler: imagehandler.NewMock(fileConfig.BaseURL, spacesClient, db),
		}

		{ // GetImage Route
			view := framework.New(imageView.GetImage)
			view.AddTenancy("testing", mockImageView.GetImage)
			server.Path("GET", "/images", view.Route)
		}

		{ // GetRandomImage Route
			view := framework.New(imageView.GetRandomImage)
			view.AddTenancy("testing", mockImageView.GetRandomImage)
			server.Path("GET", "/images/random", view.Route)
		}

		{ // GetImageByID Route
			view := framework.New(imageView.GetImageByID)
			view.AddTenancy("testing", mockImageView.GetImageByID)
			server.Path("GET", "/images/{id:\\d+}", view.Route)
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
