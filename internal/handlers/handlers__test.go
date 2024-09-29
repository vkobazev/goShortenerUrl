package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateShortURL(t *testing.T) {

	type wantResult struct {
		contentType  string
		statusCode   int
		responseBody string
		location     string
	}

	tests := []struct {
		testName    string
		httpMethod  string
		requestPath string
		requestBody string
		testUrls    map[string]string
		wantResult  wantResult
	}{
		{
			testName:    "POST request - create short URL",
			httpMethod:  http.MethodPost,
			requestPath: "/",
			requestBody: "https://example.com",
			testUrls:    map[string]string{},
			wantResult: wantResult{
				contentType:  "text/plain; charset=UTF-8",
				statusCode:   http.StatusCreated,
				responseBody: "http://localhost:8080/\\w{6}",
			},
		},
		{
			testName:    "POST request - empty requestBody",
			httpMethod:  http.MethodPost,
			requestPath: "/",
			requestBody: "",
			testUrls:    map[string]string{},
			wantResult: wantResult{
				contentType:  "text/plain; charset=UTF-8",
				statusCode:   http.StatusBadRequest,
				responseBody: "Body is empty",
			},
		},
		{
			testName:    "API POST request - create short URL",
			httpMethod:  http.MethodPost,
			requestPath: "/api/shorten",
			requestBody: "{\"url\":\"https://practicum.yandex.ru\"}",
			testUrls:    map[string]string{},
			wantResult: wantResult{
				contentType:  "application/json",
				statusCode:   http.StatusCreated,
				responseBody: "http://localhost:8080/\\w{6}",
			},
		},
		{
			testName:    "API POST request - empty requestBody",
			httpMethod:  http.MethodPost,
			requestPath: "/api/shorten",
			requestBody: "",
			testUrls:    map[string]string{},
			wantResult: wantResult{
				contentType:  "text/plain; charset=UTF-8",
				statusCode:   http.StatusBadRequest,
				responseBody: "Body is empty",
			},
		},
		{
			testName:    "GET request - valid short URL",
			httpMethod:  http.MethodGet,
			requestPath: "/abc123",
			requestBody: "",
			testUrls:    map[string]string{"abc123": "https://example.com"},
			wantResult: wantResult{
				contentType: "text/plain; charset=UTF-8",
				statusCode:  http.StatusTemporaryRedirect,
				location:    "https://example.com",
			},
		},
		{
			testName:    "GET request - invalid short URL",
			httpMethod:  http.MethodGet,
			requestPath: "/invalid",
			requestBody: "",
			testUrls:    map[string]string{},
			wantResult: wantResult{
				contentType:  "text/plain; charset=UTF-8",
				statusCode:   http.StatusNotFound,
				responseBody: "Short URL not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			// Создаем новый экземпляр Echo
			e := echo.New()

			// Устанавливаем глобальную переменную ShortList
			sh := ShortList{
				0,
				tt.testUrls,
			}

			// Регистрируем обработчики
			e.POST("/", sh.CreateShortURL)
			e.GET("/:id", sh.GetLongURL)
			e.POST("/api/shorten", sh.APIReturnShortURL)

			// Создаем тестовый сервер
			server := httptest.NewServer(e)
			defer server.Close()

			// Формируем URL для запроса
			url := server.URL + tt.requestPath

			// Создаем HTTP-клиент
			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}

			// Отправляем запрос
			var resp *http.Response
			var err error
			if tt.httpMethod == http.MethodPost {
				if tt.requestPath == "/api/shorten" {
					resp, err = client.Post(url, "application/json", strings.NewReader(tt.requestBody))

					require.NoError(t, err)
					defer resp.Body.Close()
				} else {
					resp, err = client.Post(url, "text/plain", strings.NewReader(tt.requestBody))

					require.NoError(t, err)
					defer resp.Body.Close()
				}
			} else {
				resp, err = client.Get(url)

				require.NoError(t, err)
				defer resp.Body.Close()
			}

			// Проверяем статус-код
			assert.Equal(t, tt.wantResult.statusCode, resp.StatusCode)

			// Проверяем тип контента
			assert.Equal(t, tt.wantResult.contentType, resp.Header.Get("Content-Type"))

			// Проверяем заголовок Location, если ожидается
			if tt.wantResult.location != "" {
				assert.Equal(t, tt.wantResult.location, resp.Header.Get("Location"))
			}

			// Проверяем тело ответа, если ожидается
			if tt.wantResult.responseBody != "" {
				bodyContent, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.Regexp(t, regexp.MustCompile(tt.wantResult.responseBody), string(bodyContent))
			}
		})
	}
}
