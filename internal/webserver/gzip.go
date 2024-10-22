package webserver

import (
	"bytes"
	"compress/gzip"
	"github.com/labstack/echo/v4"
	"io"
)

func DecompressGZIP(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Проверяем, что Content-Encoding - gzip
		if c.Request().Header.Get("Content-Encoding") == "gzip" {
			// Создаем gzip.Reader для декомпрессии тела запроса
			gr, err := gzip.NewReader(c.Request().Body)
			if err != nil {
				return err
			}
			defer gr.Close()

			// Читаем декомпрессированное тело в буфер
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, gr); err != nil {
				return err
			}

			// Заменяем тело запроса на декомпрессированное
			c.Request().Body = io.NopCloser(bytes.NewReader(buf.Bytes()))
		}
		return next(c)
	}
}
