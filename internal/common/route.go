package common

import (
	"github.com/julietteengel/salesrep-api/internal/utils"
	"github.com/labstack/echo/v4"
)

type ControllerType string

const (
	Public  ControllerType = "public"
	Private ControllerType = "private"
)

type CustomHandlerFunc func(c echo.Context) (error, *utils.ControllerError)

func (f CustomHandlerFunc) Handle(c echo.Context) error {
	if err, controllerError := f(c); err != nil {
		if controllerError == nil {
			controllerError = &utils.GenericServerError
		}
		return utils.WrapErrorHTTP(c, err, *controllerError)
	}
	return nil
}

type Route struct {
	Method     string
	Path       string
	Handler    CustomHandlerFunc
	Middleware []echo.MiddlewareFunc
}

type Controller interface {
	Routes() []Route
	GetType() ControllerType
}
