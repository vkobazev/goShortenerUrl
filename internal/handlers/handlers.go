package handlers

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/vkobazev/goShortenerUrl/internal/config"
	"github.com/vkobazev/goShortenerUrl/internal/consts"
	"github.com/vkobazev/goShortenerUrl/internal/data"
	"github.com/vkobazev/goShortenerUrl/internal/database"
	jwt "github.com/vkobazev/goShortenerUrl/internal/jwt"
	"io"
	"log"
	"math/rand"
	"net/http"
)

type URLShortener struct {
	Counter uint
	URLS    map[string]string
	ReURLS  map[string]string
	Tests   bool
	DB      *database.DB
}

type ShortResponse struct {
	Result string `json:"result"`
	UserID string `json:"user_id,omitempty"`
}

type LongResponse struct {
	ID       string `json:"correlation_id"`
	ShortURL string `json:"short_url"`
	UserID   string `json:"user_id,omitempty"`
}

func NewShortList() *URLShortener {
	return &URLShortener{
		Counter: 0,
		URLS:    make(map[string]string),
		ReURLS:  make(map[string]string),
		Tests:   false,
	}
}

func (sh *URLShortener) CreateShortURL(c echo.Context) error {

	userID := c.Get(jwt.UserIDKey).(string)

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(http.StatusBadRequest, "Read Body failed")
	}
	if len(body) == 0 {
		return c.String(http.StatusBadRequest, "Body is empty")
	}

	longURL := string(body)
	shortURL, err := sh.StoreURL(longURL, userID)
	if err != nil {
		if err.Error() == "conflict" {
			c.Response().Header().Set("Content-Type", "text/plain; charset=UTF-8")
			c.Response().WriteHeader(http.StatusConflict)
			return c.String(http.StatusConflict, shortURL)
		}
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	c.Response().Header().Set("Content-Type", "text/plain; charset=UTF-8")
	c.Response().WriteHeader(http.StatusCreated)
	return c.String(http.StatusCreated, shortURL)
}

func (sh *URLShortener) GetLongURL(c echo.Context) error {

	id := c.Param("id")

	longURL, _, err, deleted := sh.RetrieveURL(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Short URL not found")
	}
	if deleted {
		return c.String(http.StatusGone, "410 Gone")
	}

	c.Response().Header().Set("Content-Type", "text/plain; charset=UTF-8")
	return c.Redirect(http.StatusTemporaryRedirect, longURL)
}

func (sh *URLShortener) APIReturnShortURL(c echo.Context) error {

	userID := c.Get(jwt.UserIDKey).(string)

	var requestData struct {
		URL string `json:"url"`
	}

	if err := c.Bind(&requestData); err != nil {
		return c.String(http.StatusBadRequest, "Read Body failed")
	}
	if requestData.URL == "" {
		return c.String(http.StatusBadRequest, "Body is empty")
	}

	shortURL, err := sh.StoreURL(requestData.URL, userID)
	if err != nil {
		if err.Error() == "conflict" {
			response := ShortResponse{
				Result: shortURL,
				UserID: userID,
			}
			return c.JSON(http.StatusConflict, response)
		}
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	response := ShortResponse{
		Result: shortURL,
		UserID: userID,
	}
	return c.JSON(http.StatusCreated, response)
}

func (sh *URLShortener) APIPutMassiveData(c echo.Context) error {

	userID := c.Get(jwt.UserIDKey).(string)

	var requestDataSlice []database.RequestData
	if err := c.Bind(&requestDataSlice); err != nil {
		return c.String(http.StatusBadRequest, "Read Body failed")
	}

	// Add userID to each request
	for i := range requestDataSlice {
		requestDataSlice[i].UserID = userID
	}

	response, err := sh.StoreURLBatch(requestDataSlice)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	return c.JSON(http.StatusCreated, response)
}

func (sh *URLShortener) APIReturnUserData(c echo.Context) error {

	userID := c.Get(jwt.UserIDKey).(string)

	urls, err := sh.DB.GetURLsByUser(context.Background(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Failed to retrieve user URLs",
		})
	}
	if urls == nil {
		return c.JSON(http.StatusUnauthorized, urls)
	}

	return c.JSON(http.StatusOK, urls)
}

func (sh *URLShortener) APIDeleteUserURLs(c echo.Context) error {
	userID := c.Get(jwt.UserIDKey).(string)

	var shortURLs []string
	if err := c.Bind(&shortURLs); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request body")
	}

	if len(shortURLs) == 0 {
		return c.NoContent(http.StatusAccepted)
	}

	ctx := context.Background()
	const batchSize = 1000
	const numWorkers = 5

	batch := func(urls []string, size int) [][]string {
		var batches [][]string
		for size < len(urls) {
			urls, batches = urls[size:], append(batches, urls[0:size:size])
		}
		if len(urls) > 0 {
			batches = append(batches, urls)
		}
		return batches
	}

	batches := batch(shortURLs, batchSize)
	jobs := make(chan []string, len(batches))
	results := make(chan error, len(batches))

	for w := 0; w < numWorkers; w++ {
		go func() {
			for batch := range jobs {
				err := sh.DB.DeleteURLforUser(ctx, userID, batch)
				results <- err
			}
		}()
	}

	go func() {
		for _, batch := range batches {
			jobs <- batch
		}
		close(jobs)
	}()

	return c.NoContent(http.StatusAccepted)
}

// Helper functions

func (sh *URLShortener) PingDB(c echo.Context) error {
	err := sh.DB.Ping(context.Background())
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

// Business logic functions

func (sh *URLShortener) StoreURL(longURL, userID string) (string, error) {
	id := GenRandomID(consts.ShortURLLength)
	host := config.Options.ReturnAddr
	if host == "" {
		host = consts.HTTPMethod + "://" + "localhost:8080"
	}
	shortURL := host + "/" + id

	switch {
	case config.Options.DataBaseConn == "":
		oldID, ok := sh.ReURLS[longURL]
		if !ok {
			sh.URLS[id] = longURL
			sh.ReURLS[longURL] = id
			sh.Counter++

			if !sh.Tests {
				err := data.P.WriteEvent(&data.Event{
					ID:     sh.Counter,
					Short:  id,
					Long:   longURL,
					UserID: userID,
				})
				if err != nil {
					log.Fatalf("Error writing Event: %v", err)
				}
			}
			return shortURL, nil
		} else {
			shortURL = host + "/" + oldID
			return shortURL, fmt.Errorf("conflict")
		}
	default:
		exists, err := sh.DB.LongURLExists(context.Background(), longURL, userID)
		if err != nil {
			log.Fatalf("Error checking long URL existence: %v", err)
		}
		if exists {
			oldID, err := sh.DB.GetShortURL(context.Background(), longURL, userID)
			if err != nil {
				return "", err
			}
			shortURL = host + "/" + oldID
			return shortURL, fmt.Errorf("conflict")
		} else {
			err = sh.DB.InsertURL(context.Background(), id, longURL, userID)
			if err != nil {
				log.Fatalf("Error inserting URL: %v", err)
			}
			return shortURL, nil
		}
	}
}

func (sh *URLShortener) RetrieveURL(id string) (string, string, error, bool) {
	switch {
	case config.Options.DataBaseConn == "":
		longURL, ok := sh.URLS[id]
		if !ok {
			return "", "", fmt.Errorf("not found"), false
		}
		return longURL, "", nil, false
	default:
		_, deleted, err := sh.DB.LongURLDeleted(context.Background(), id)
		if err != nil {
			return "", "", err, false
		}
		if deleted {
			return "", "", err, deleted
		}

		longURL, userID, err := sh.DB.GetLongURL(context.Background(), id)
		if err != nil {
			return "", "", err, false
		}
		return longURL, userID, nil, false
	}
}

func (sh *URLShortener) StoreURLBatch(requestDataSlice []database.RequestData) ([]LongResponse, error) {
	host := config.Options.ReturnAddr
	if host == "" {
		host = consts.HTTPMethod + "://" + "localhost:8080"
	}

	switch {
	case config.Options.DataBaseConn == "":
		for _, pair := range requestDataSlice {
			sh.URLS[pair.ID] = pair.URL
			sh.ReURLS[pair.URL] = pair.ID
			sh.Counter++

			// Event writing
			if !sh.Tests {
				err := data.P.WriteEvent(&data.Event{
					ID:    sh.Counter,
					Short: pair.ID,
					Long:  pair.URL,
				})
				if err != nil {
					log.Fatalf("Error writing Event: %v", err)
				}
			}
		}
	default:
		err := sh.DB.InsertURLs(context.Background(), requestDataSlice)
		if err != nil {
			log.Fatalf("Error inserting URLs: %v", err)
		}
	}

	var response []LongResponse
	for _, pair := range requestDataSlice {
		long := LongResponse{
			ID:       pair.ID,
			ShortURL: host + "/" + pair.ID,
		}
		response = append(response, long)
	}

	return response, nil
}
