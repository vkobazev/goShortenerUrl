package webserver

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/vkobazev/goShortenerUrl/config"
	"github.com/vkobazev/goShortenerUrl/handlers"
)

func WebServer() {

	e := echo.New()
	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Create and return the group
	g := e.Group("/")
	{
		// Define routes
		g.POST("", handlers.CreateShortURL)
		g.GET(":id", handlers.GetLongURL)
	}

	e.Logger.Fatal(e.Start(config.Options.ListenAddr))
}
