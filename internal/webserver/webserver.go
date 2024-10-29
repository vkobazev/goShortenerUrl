package webserver

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/vkobazev/goShortenerUrl/internal/config"
	"github.com/vkobazev/goShortenerUrl/internal/data"
	"github.com/vkobazev/goShortenerUrl/internal/database"
	"github.com/vkobazev/goShortenerUrl/internal/handlers"
	"github.com/vkobazev/goShortenerUrl/internal/jwt"
	"github.com/vkobazev/goShortenerUrl/internal/logger"
	"go.uber.org/zap"
	"log"
)

func StartWebServer() {
	// Create New map for Short links list
	sh := handlers.NewShortList()

	if config.Options.DataBaseConn != "" {
		InitDB(sh)
		defer sh.DB.Close()
	} else {
		SetupEvents(sh)
		defer data.P.Close()
	}

	l := SetupLogger()
	SetupEcho(l, sh)
}

func SetupLogger() *zap.Logger {
	// Create logger struct
	l, err := logger.InitLogger("./shortener.log")
	if err != nil {
		log.Fatalf("failed init logger %s", err)
	}
	return l
}

func SetupEvents(sh *handlers.URLShortener) {
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
}

func SetupEcho(l *zap.Logger, sh *handlers.URLShortener) {
	e := echo.New()
	// Add middleware
	e.Use(middleware.Logger())
	e.Use(logger.LoggerMiddleware(l))

	e.Use(DecompressGZIP) // Gzip middlewares
	e.Use(middleware.Gzip())
	e.Use(jwt.JWTMiddleware())

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
			api.POST("shorten/batch", sh.APIPutMassiveData)

			user := api.Group("user/")
			{
				user.GET("urls", sh.APIReturnUserData)
				user.DELETE("urls", sh.APIDeleteUserURLs)
			}
		}
	}

	e.Logger.Fatal(e.Start(config.Options.ListenAddr))
}

func InitDB(sh *handlers.URLShortener) {
	// Init DB
	var err error

	sh.DB, err = database.New(config.Options.DataBaseConn)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	err = sh.DB.CreateTable(context.Background())
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}
}
