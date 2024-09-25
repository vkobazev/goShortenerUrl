package logger

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"time"
)

func InitLogger() *zap.Logger {
	logger := zap.Must(zap.NewProduction())

	defer logger.Sync()

	return logger
}

func LoggerMiddleware(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			w := c.Request()
			r := c.Response()
			start := time.Now()

			err := next(c)

			stop := time.Now()
			latency := stop.Sub(start)

			ip := c.RealIP()
			method := w.Method
			uri := w.RequestURI
			status := r.Status
			userAgent := w.UserAgent()
			referer := w.Referer()

			logger.Info("HTTP request",
				zap.String("ip", ip),
				zap.String("method", method),
				zap.String("uri", uri),
				zap.Int("status", status),
				zap.String("user_agent", userAgent),
				zap.String("referer", referer),
				zap.Duration("latency", latency),
			)

			return err
		}
	}
}
