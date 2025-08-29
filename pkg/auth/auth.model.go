package auth

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"strings"
)

type BaseAuthData struct {
	UserID    uint
	CompanyID *uint
	Status    string
}

type AuthData struct {
	BaseAuthData
	Role      string
	ManagerID *uint
}

// CustomClaims contains custom data we want from the token.
type CustomClaims struct {
	jwt.RegisteredClaims
	Scope         string   `json:"scope"`
	Audience      []string `json:"aud"`
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
}

func CheckClaims(scopes string, requiredScope string) error {
	contains := strings.Contains(scopes, requiredScope)
	if !contains {
		return errors.New("missing scope")
	}
	return nil
}
