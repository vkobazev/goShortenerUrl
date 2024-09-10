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

	type want struct {
		contentType  string
		statusCode   int
		responseBody string
		location     string
	}

	tests := []struct {
		name   string
		method string
		path   string
		body   string
		urls   map[string]string
		want   want
	}{
		{
			name:   "POST request - create short URL",
			method: http.MethodPost,
			path:   "/",
			body:   "https://example.com",
			urls:   map[string]string{},
			want: want{
				contentType:  "text/plain; charset=UTF-8",
				statusCode:   http.StatusCreated,
				responseBody: "http://localhost:8080/\\w{6}",
			},
		},
		{
			name:   "POST request - empty body",
			method: http.MethodPost,
			path:   "/",
			body:   "",
			urls:   map[string]string{},
			want: want{
				contentType:  "text/plain; charset=UTF-8",
				statusCode:   http.StatusBadRequest,
				responseBody: "Body is empty",
			},
		},
		{
			name:   "GET request - valid short URL",
			method: http.MethodGet,
			path:   "/abc123",
			body:   "",
			urls:   map[string]string{"abc123": "https://example.com"},
			want: want{
				contentType: "text/plain; charset=UTF-8",
				statusCode:  http.StatusTemporaryRedirect,
				location:    "https://example.com",
			},
		},
		{
			name:   "GET request - invalid short URL",
			method: http.MethodGet,
			path:   "/invalid",
			body:   "",
			urls:   map[string]string{},
			want: want{
				contentType:  "text/plain; charset=UTF-8",
				statusCode:   http.StatusNotFound,
				responseBody: "Short URL not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем новый экземпляр Echo
			e := echo.New()

			// Устанавливаем глобальную переменную Urls
			Urls = tt.urls

			// Регистрируем обработчики
			e.POST("/", CreateShortURL)
			e.GET("/:id", GetLongURL)

			// Создаем тестовый сервер
			server := httptest.NewServer(e)
			defer server.Close()

			// Формируем URL для запроса
			url := server.URL + tt.path

			// Создаем HTTP-клиент
			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}

			// Отправляем запрос
			var resp *http.Response
			var err error
			if tt.method == http.MethodPost {
				resp, err = client.Post(url, "text/plain", strings.NewReader(tt.body))

				require.NoError(t, err)
				defer resp.Body.Close()
			} else {
				resp, err = client.Get(url)

				require.NoError(t, err)
				defer resp.Body.Close()
			}

			// Проверяем статус-код
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

			// Проверяем тип контента
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))

			// Проверяем заголовок Location, если ожидается
			if tt.want.location != "" {
				assert.Equal(t, tt.want.location, resp.Header.Get("Location"))
			}

			// Проверяем тело ответа, если ожидается
			if tt.want.responseBody != "" {
				bodyContent, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.Regexp(t, regexp.MustCompile(tt.want.responseBody), string(bodyContent))
			}
		})
	}
}
