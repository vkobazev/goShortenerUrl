package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/vkobazev/goShortenerUrl/config"
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
		g.POST("", CreateShortURL)
		g.GET(":id", GetLongURL)
	}

	e.Logger.Fatal(e.Start(config.Options.Host + ":" + config.Options.Port))
}
