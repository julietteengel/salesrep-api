package main

import (
	_ "github.com/julietteengel/salesrep-api/docs"
	"github.com/julietteengel/salesrep-api/internal/common"
	customMiddleware "github.com/julietteengel/salesrep-api/pkg/middleware"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/swaggo/echo-swagger"
	"gorm.io/gorm"

	"net/http"
)

func NewEchoServer(controllers []common.Controller, db *gorm.DB) *echo.Echo {
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

	// Protected API routes with Auth0
	api := e.Group("/api")
	api.Use(customMiddleware.Auth0Middleware())
	api.Use(customMiddleware.SyncUserMiddleware(db))

	for _, ctrl := range controllers {
		switch ctrl.GetType() {
		case common.Private:
			for _, route := range ctrl.Routes() {
				api.Add(route.Method, route.Path, route.Handler.Handle, route.Middleware...)
			}
		}
	}

	return e
}
