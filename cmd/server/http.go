package main

import (
	_ "github.com/julietteengel/salesrep-api/docs"
	"github.com/julietteengel/salesrep-api/internal/common"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/swaggo/echo-swagger"
	"net/http"
)

func NewEchoServer(controllers []common.Controller) *echo.Echo {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// Swagger secured route
	swagger := e.Group("/swagger", middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if username == "admin" && password == "mypassword" {
			return true, nil
		}
		return false, nil
	}))
	swagger.GET("/*", echoSwagger.WrapHandler)

	// Private group
	privateApi := e.Group("/api/v1")

	for _, ctrl := range controllers {
		switch ctrl.GetType() {
		case common.Private:
			for _, route := range ctrl.Routes() {
				privateApi.Add(route.Method, route.Path, route.Handler, route.Middleware...)
			}
		}
	}

	return e
}
