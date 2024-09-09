package main

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
	baseURL := "http://localhost:8080/"

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
			path:   baseURL,
			body:   "https://example.com",
			urls:   map[string]string{},
			want: want{
				contentType:  "text/plain; charset=UTF-8",
				statusCode:   http.StatusCreated,
				responseBody: baseURL + "\\w{6}",
			},
		},
		{
			name:   "POST request - empty body",
			method: http.MethodPost,
			path:   baseURL,
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
			path:   baseURL + "invalid",
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
			e := echo.New()
			Urls = tt.urls

			request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			c := e.NewContext(request, w)

			var err error
			if tt.method == http.MethodPost {
				err = CreateShortURL(c)
			} else {
				c.SetParamNames("id")
				c.SetParamValues(tt.path[1:])

				err = GetLongURL(c)
			}
			require.NoError(t, err)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			if tt.want.location != "" {
				assert.Equal(t, tt.want.location, result.Header.Get("Location"))
			}

			if tt.want.responseBody != "" {
				bodyContent, err := io.ReadAll(result.Body)
				require.NoError(t, err)
				assert.Regexp(t, regexp.MustCompile(tt.want.responseBody), string(bodyContent))
			}
		})
	}
}
