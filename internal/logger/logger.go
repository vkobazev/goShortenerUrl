package logger

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"time"
)

func InitLogger(logFile string) (*zap.Logger, error) {

	// Create Log rotation
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    500,
		MaxBackups: 3,
		MaxAge:     28,
	})

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Wrap Lumberjack to Zap configuration
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		w,
		zap.InfoLevel,
	)

	logger := zap.New(core, zap.AddCaller())

	return logger, nil
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
