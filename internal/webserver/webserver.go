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

	e := echo.New()
	// Create New map for Short links list
	sh := handlers.NewShortList()

	if config.Options.DataBaseConn != "" {
		// Init DB
		var err error

		sh.DB, err = database.New(config.Options.DataBaseConn)
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer sh.DB.Close(context.Background())

		err = sh.DB.CreateTable(context.Background())
		if err != nil {
			log.Fatalf("Error creating table: %v", err)
		}
	} else {
		// New Consumer to restore data
		C, err := data.NewConsumer(config.Options.FileStoragePath)
		if err != nil {
			log.Fatalf("Error creating consumer for Events: %v", err)
		}
		events, err := C.ReadAllEvents()
		if err != nil {
			log.Fatalf("Error restore DATA from Events: %v", err)
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
	}

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
		g.GET("ping", sh.PingDB)

		// Define api group
		api := g.Group("api/")
		{
			api.POST("shorten", sh.APIReturnShortURL)
		}
	}

	e.Logger.Fatal(e.Start(config.Options.ListenAddr))
}
