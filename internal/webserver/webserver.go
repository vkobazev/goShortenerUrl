package webserver

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/vkobazev/goShortenerUrl/internal/config"
	"github.com/vkobazev/goShortenerUrl/internal/handlers"
	"github.com/vkobazev/goShortenerUrl/internal/logger"
)

func WebServer() {

	e := echo.New()
	// Create New map for Short links list
	sh := handlers.NewShortList()

	// Create logger struct
	l := logger.InitLogger()

	// Add middleware
	e.Use(middleware.Logger())
	e.Use(logger.LoggerMiddleware(l))

	// Create and return the group
	g := e.Group("/")
	{
		// Define routes
		g.POST("", sh.CreateShortURL)
		g.GET(":id", sh.GetLongURL)
	}

	e.Logger.Fatal(e.Start(config.Options.ListenAddr))
}
