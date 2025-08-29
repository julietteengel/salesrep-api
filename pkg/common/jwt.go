package common

import (
	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	jwt.RegisteredClaims
	Scope         string   `json:"scope"`
	Audience      []string `json:"aud,omitempty"`
	Email         string   `json:"https://salesrep.com/email,omitempty"`
	EmailVerified bool     `json:"https://salesrep.com/email_verified,omitempty"`
	Name          string   `json:"https://salesrep.com/name,omitempty"`
	Picture       string   `json:"https://salesrep.com/picture,omitempty"`
	Auth0ID       string   `json:"sub"`
}

type UserContext struct {
	ID            uint   `json:"id"`
	Auth0ID       string `json:"auth0_id"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Role          string `json:"role"`
	IsOnboarded   bool   `json:"is_onboarded"`
}