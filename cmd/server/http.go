package main

import (
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/julietteengel/salesrep-api/docs"
	"github.com/julietteengel/salesrep-api/internal/common"
	"github.com/julietteengel/salesrep-api/pkg/auth"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
	"github.com/swaggo/echo-swagger"

	"net/http"
)

func NewEchoServer(controllers []common.Controller, authService auth.IAuthService) *echo.Echo {
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
	privateApi := e.Group("/v0")
	privateApi.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(viper.GetString("JWT_KEY")),
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return &auth.CustomClaims{}
		},
	}))
	privateApi.Use(UserVerifiedMiddleware)
	privateApi.Use(UserMiddleware(authService))

	for _, ctrl := range controllers {
		switch ctrl.GetType() {
		case common.Private:
			for _, route := range ctrl.Routes() {
				privateApi.Add(route.Method, route.Path, route.Handler.Handle, route.Middleware...)
			}
		}
	}

	return e
}
