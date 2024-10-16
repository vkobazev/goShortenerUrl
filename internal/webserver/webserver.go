package webserver

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/vkobazev/goShortenerUrl/internal/config"
	"github.com/vkobazev/goShortenerUrl/internal/data"
	"github.com/vkobazev/goShortenerUrl/internal/database"
	"github.com/vkobazev/goShortenerUrl/internal/handlers"
	"github.com/vkobazev/goShortenerUrl/internal/logger"
	"log"
)

func WebServer() {

	// Init DB
	//conn := database.PostgresConfig{
	//	Host:     "localhost",
	//	Port:     5432,
	//	User:     "postgres",
	//	Password: "1234",
	//	DBName:   "postgres",
	//}

	db, err := database.NewDB()
	if err != nil {
		// Handle the error
		log.Fatal(err)
	}
	defer db.Close(context.Background())

	e := echo.New()
	// Create New map for Short links list
	sh := handlers.NewShortList()

	// New Consumer to restore data
	C, err := data.NewConsumer(config.Options.FileStoragePath)
	if err != nil {
		panic(err)
	}
	events, err := C.ReadAllEvents()
	if err != nil {
		panic(err)
	}
	for _, event := range events {
		sh.URLS[event.Short] = event.Long
		sh.Counter = event.ID
	}

	// New Event producer
	data.P, err = data.NewProducer(config.Options.FileStoragePath)
	if err != nil {
		panic(err)
	}
	defer data.P.Close()

	// Create logger struct
	l, err := logger.InitLogger("./shortener.log")
	if err != nil {
		log.Fatalf("failed init logger %s", err)
	}

	// Add middleware
	e.Use(middleware.Logger())
	e.Use(logger.LoggerMiddleware(l))

	e.Use(DecompressGZIP) // Gzip middlewares
	e.Use(middleware.Gzip())

	// Create and return the group
	g := e.Group("/")
	{
		// Define routes
		g.POST("", sh.CreateShortURL)
		g.GET(":id", sh.GetLongURL)
		g.GET("ping", func(c echo.Context) error {
			return handlers.PingDB(c, db)
		})

		// Define api group
		api := g.Group("api/")
		{
			api.POST("shorten", sh.APIReturnShortURL)
		}
	}

	e.Logger.Fatal(e.Start(config.Options.ListenAddr))
}
