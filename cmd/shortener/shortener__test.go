package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestUrlShortener(t *testing.T) {
	baseURL := "http://localhost:8080"
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
			path:   baseURL + "/",
			body:   "https://example.com",
			urls:   map[string]string{},
			want: want{
				contentType:  "text/plain; charset=utf-8",
				statusCode:   http.StatusCreated,
				responseBody: baseURL + "/" + "\\w{6}",
			},
		},
		{
			name:   "GET request - valid short URL",
			method: http.MethodGet,
			path:   baseURL + "/abc123",
			body:   "",
			urls: map[string]string{
				"/abc123": "https://example.com",
			},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusTemporaryRedirect,
				location:    "https://example.com",
			},
		},
		{
			name:   "GET request - invalid short URL",
			method: http.MethodGet,
			path:   baseURL + "/invalid",
			body:   "",
			urls:   map[string]string{},
			want: want{
				contentType:  "text/plain; charset=utf-8",
				statusCode:   http.StatusNotFound,
				responseBody: "Short URL not found",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Urls = tt.urls
			request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			urlShortener(w, request)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			urlResult, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			if tt.want.responseBody != "" {
				assert.Regexp(t, regexp.MustCompile(tt.want.responseBody), string(urlResult))
			}
		})

	}
}
