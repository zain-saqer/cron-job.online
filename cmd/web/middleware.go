package main

import (
	"crypto/subtle"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"os"
)

func requestLogger(e *echo.Echo) {
	logger := zerolog.New(os.Stdout)
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.Info().
				Str("URI", v.URI).
				Int("status", v.Status).
				Msg("request")

			return nil
		},
	}))
}

func basicAuth(e *echo.Echo, user, pass string) {
	e.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if subtle.ConstantTimeCompare([]byte(username), []byte(user)) == 1 &&
			subtle.ConstantTimeCompare([]byte(password), []byte(pass)) == 1 {
			return true, nil
		}
		return false, nil
	}))
}

func middlewares(e *echo.Echo, config *Config) {
	e.Use(middleware.Recover())
	basicAuth(e, config.AuthUser, config.AuthPass)
	e.Use(middleware.Gzip())
	requestLogger(e)
	e.Use(sentryecho.New(sentryecho.Options{
		Repanic: true,
	}))
}
