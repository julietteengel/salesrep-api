package main

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/julietteengel/salesrep-api/pkg/auth"
	"github.com/labstack/echo/v4"
)

func UserVerifiedMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Handle email not verified
		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(*auth.CustomClaims)
		if !claims.EmailVerified {
			return echo.NewHTTPError(403, "Email is not verified")
		}
		return next(c)
	}
}

type ctxKey string

const (
	LocaleKey    ctxKey = "locale"
	UserIDKey    ctxKey = "userID"
	CompanyIDKey ctxKey = "companyID"
)

func UserMiddleware(authService auth.IAuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Handle email not verified
			user := c.Get("user").(*jwt.Token)
			claims := user.Claims.(*auth.CustomClaims)
			// Handle user not found
			authData, err := authService.GetUserAuthData(c.Request().Context(), claims.Subject)
			if err != nil {
				return echo.NewHTTPError(401, "User not found")
			}
			c.Set("authData", authData)
			goContext := context.WithValue(c.Request().Context(), UserIDKey, authData.UserID)
			c.SetRequest(c.Request().WithContext(goContext))
			return next(c)
		}
	}
}
