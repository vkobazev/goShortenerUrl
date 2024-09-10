package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/vkobazev/goShortenerUrl/config"
	"io"
	"math/rand"
	"net/http"
)

var (
	Urls = map[string]string{}
)

func CreateShortURL(c echo.Context) error {
	// Handle POST request
	c.Request().URL.Scheme = "http"
	id := RandomString(6)
	host := config.Options.ReturnAddr
	if host == "" {
		host = "localhost:8080"
	}
	shortURL := c.Request().URL.Scheme + "://" + host + "/" + id

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

// Helper func for token generate

func RandomString(num int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	str := make([]rune, num)
	for i := range str {
		str[i] = letters[rand.Intn(len(letters))]
	}
	return string(str)
}
