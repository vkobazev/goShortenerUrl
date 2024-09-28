package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/vkobazev/goShortenerUrl/internal/config"
	"github.com/vkobazev/goShortenerUrl/internal/consts"
	"io"
	"math/rand"
	"net/http"
)

type ShortList struct {
	urls map[string]string
}

type ShortResponse struct {
	Result string `json:"result"`
}

func NewShortList() *ShortList {
	return &ShortList{urls: make(map[string]string)}
}

func (sh *ShortList) CreateShortURL(c echo.Context) error {
	// Handle POST request
	id := GenRandomID(consts.ShortURLLength)
	host := config.Options.ReturnAddr
	if host == "" {
		host = consts.HTTPMethod + "://" + "localhost:8080"
	}
	shortURL := host + "/" + id

	// Read requestBody to create Map
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(http.StatusBadRequest, "Read Body failed")
	}
	if string(body) == "" {
		return c.String(http.StatusBadRequest, "Body is empty")
	}
	sh.urls[id] = string(body)

	// Response writing
	c.Response().Header().Set("Content-Type", "text/plain; charset=UTF-8")
	c.Response().WriteHeader(http.StatusCreated)

	return c.String(http.StatusCreated, shortURL)
}

func (sh *ShortList) GetLongURL(c echo.Context) error {
	// Handle GET request
	id := c.Param("id")
	long, ok := sh.urls[id]
	if !ok {
		return c.String(http.StatusNotFound, "Short URL not found")
	}
	// Response writing
	c.Response().Header().Set("Content-Type", "text/plain; charset=UTF-8")
	return c.Redirect(http.StatusTemporaryRedirect, long)
}

func (sh *ShortList) APIReturnShortURL(c echo.Context) error {
	// Assuming you have a struct to decode the request body
	var requestData struct {
		URL string `json:"url"`
	}

	if err := c.Bind(&requestData); err != nil {
		return c.String(http.StatusBadRequest, "Read Body failed")
	}
	if requestData.URL == "" {
		return c.String(http.StatusBadRequest, "Body is empty")
	}

	id := GenRandomID(consts.ShortURLLength)
	host := config.Options.ReturnAddr
	if host == "" {
		host = consts.HTTPMethod + "://" + "localhost:8080"
	}

	sh.urls[id] = requestData.URL

	response := ShortResponse{
		Result: host + "/" + id,
	}

	return c.JSON(http.StatusCreated, response)
}

// Helper func for token generate

func GenRandomID(num int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	str := make([]rune, num)
	for i := range str {
		str[i] = letters[rand.Intn(len(letters))]
	}
	return string(str)
}
