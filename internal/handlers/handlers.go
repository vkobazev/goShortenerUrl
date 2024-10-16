package handlers

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/vkobazev/goShortenerUrl/internal/config"
	"github.com/vkobazev/goShortenerUrl/internal/consts"
	"github.com/vkobazev/goShortenerUrl/internal/data"
	"io"
	"math/rand"
	"net/http"
)

type ShortList struct {
	Counter uint
	URLS    map[string]string
	tests   bool
}

type ShortResponse struct {
	Result string `json:"result"`
}

func NewShortList() *ShortList {
	return &ShortList{
		Counter: 0,
		URLS:    make(map[string]string),
		tests:   false,
	}
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
	sh.URLS[id] = string(body)
	sh.Counter++

	// Event writing
	if !sh.tests {
		err = data.P.WriteEvent(&data.Event{
			ID:    sh.Counter,
			Short: id,
			Long:  string(body),
		})
		if err != nil {
			panic(err)
		}
	}

	// Response writing
	c.Response().Header().Set("Content-Type", "text/plain; charset=UTF-8")
	c.Response().WriteHeader(http.StatusCreated)

	return c.String(http.StatusCreated, shortURL)
}

func (sh *ShortList) GetLongURL(c echo.Context) error {
	// Handle GET request
	id := c.Param("id")
	long, ok := sh.URLS[id]
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

	sh.URLS[id] = requestData.URL
	sh.Counter++

	// Event writing
	if !sh.tests {
		err := data.P.WriteEvent(&data.Event{
			ID:    sh.Counter,
			Short: id,
			Long:  requestData.URL,
		})
		if err != nil {
			panic(err)
		}
	}

	response := ShortResponse{
		Result: host + "/" + id,
	}

	return c.JSON(http.StatusCreated, response)
}

// Helper func

func PingDB(c echo.Context, db *pgx.Conn) error {
	err := db.Ping(context.Background())
	if err != nil {
		return c.String(http.StatusInternalServerError, "500 Internal Server Error")
	}

	// Response writing
	c.Response().Header().Set("Content-Type", "text/plain; charset=UTF-8")
	return c.String(http.StatusOK, "OK")
}

func GenRandomID(num int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	str := make([]rune, num)
	for i := range str {
		str[i] = letters[rand.Intn(len(letters))]
	}
	return string(str)
}
