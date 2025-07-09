package common

import "github.com/labstack/echo/v4"

type ControllerType string

const (
	Public  ControllerType = "public"
	Private ControllerType = "private"
)

type Route struct {
	Method     string
	Path       string
	Handler    echo.HandlerFunc
	Middleware []echo.MiddlewareFunc
}

type Controller interface {
	Routes() []Route
	GetType() ControllerType
}
