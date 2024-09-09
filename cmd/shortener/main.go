package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"io"
	"math/rand"
	"net/http"
)

var (
	Urls = map[string]string{}
)

func main() {
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

	e.Logger.Fatal(e.Start(":8080"))
}

func CreateShortURL(c echo.Context) error {
	// Handle POST request
	c.Request().URL.Scheme = "http"
	id := RandomString(6)
	shortURL := c.Request().URL.Scheme + "://" + c.Request().Host + "/" + id

	// Read body to create Map
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(http.StatusBadRequest, "Read Body failed")
	}
	if string(body) == "" {
		return c.String(http.StatusBadRequest, "Body is empty")
	}
	Urls[id] = string(body)

	// Response writing
	c.Response().Header().Set("Content-Type", "text/plain; charset=UTF-8")
	c.Response().WriteHeader(http.StatusCreated)

	return c.String(http.StatusCreated, shortURL)
}

func GetLongURL(c echo.Context) error {
	// Handle GET request
	id := c.Param("id")
	long, ok := Urls[id]
	if !ok {
		return c.String(http.StatusNotFound, "Short URL not found")
	}
	// Response writing
	c.Response().Header().Set("Content-Type", "text/plain; charset=UTF-8")
	return c.Redirect(http.StatusTemporaryRedirect, long)
}

func RandomString(num int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	str := make([]rune, num)
	for i := range str {
		str[i] = letters[rand.Intn(len(letters))]
	}
	return string(str)
}
