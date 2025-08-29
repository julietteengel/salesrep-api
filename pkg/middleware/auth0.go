package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/julietteengel/salesrep-api/pkg/user"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type Auth0Claims struct {
	jwt.RegisteredClaims
	Scope         string   `json:"scope"`
	Permissions   []string `json:"permissions,omitempty"`
	Email         string   `json:"email,omitempty"`
	EmailVerified bool     `json:"email_verified,omitempty"`
}

type JWKSResponse struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kid string   `json:"kid"`
	Kty string   `json:"kty"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5C []string `json:"x5c,omitempty"`
}

var jwksCache *JWKSResponse

func getJWKS() (*JWKSResponse, error) {
	if jwksCache != nil {
		return jwksCache, nil
	}

	domain := viper.GetString("AUTH0_DOMAIN")
	resp, err := http.Get(fmt.Sprintf("https://%s/.well-known/jwks.json", domain))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var jwks JWKSResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}

	jwksCache = &jwks
	return &jwks, nil
}

func Auth0Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

			audience := viper.GetString("AUTH0_AUDIENCE")
			issuer := fmt.Sprintf("https://%s/", viper.GetString("AUTH0_DOMAIN"))

			token, err := jwt.ParseWithClaims(tokenString, &Auth0Claims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}

				kid, ok := token.Header["kid"].(string)
				if !ok {
					return nil, errors.New("kid not found in token header")
				}

				jwks, err := getJWKS()
				if err != nil {
					return nil, err
				}

				for _, key := range jwks.Keys {
					if key.Kid == kid {
						cert := "-----BEGIN CERTIFICATE-----\n" + key.X5C[0] + "\n-----END CERTIFICATE-----"
						return jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
					}
				}

				return nil, errors.New("unable to find appropriate key")
			}, jwt.WithAudience(audience), jwt.WithIssuer(issuer))

			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			if !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			claims, ok := token.Claims.(*Auth0Claims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid claims")
			}

			c.Set("auth0_claims", claims)
			c.Set("auth0_id", claims.Subject)
			c.Set("email", claims.Email)
			c.Set("email_verified", claims.EmailVerified)

			return next(c)
		}
	}
}

func RequireEmailVerified() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			emailVerified, ok := c.Get("email_verified").(bool)
			if !ok || !emailVerified {
				return echo.NewHTTPError(http.StatusForbidden, "email verification required")
			}
			return next(c)
		}
	}
}

func RequirePermission(permission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, ok := c.Get("auth0_claims").(*Auth0Claims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid claims")
			}

			for _, p := range claims.Permissions {
				if p == permission {
					return next(c)
				}
			}

			return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
		}
	}
}

func SyncUserMiddleware(db *gorm.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, ok := c.Get("auth0_claims").(*Auth0Claims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid claims")
			}

			auth0ID := claims.Subject
			email := claims.Email
			emailVerified := claims.EmailVerified

			var existingUser user.User
			result := db.Where("auth0_id = ?", auth0ID).First(&existingUser)

			if result.Error != nil {
				if errors.Is(result.Error, gorm.ErrRecordNotFound) {
					newUser := user.User{
						Auth0ID:       auth0ID,
						Email:         email,
						EmailVerified: emailVerified,
						Role:          "sales_rep",
						LastLogin:     &time.Time{},
					}

					if err := db.Create(&newUser).Error; err != nil {
						return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
					}

					c.Set("current_user", &newUser)
					c.Set("user_id", newUser.ID)
				} else {
					return echo.NewHTTPError(http.StatusInternalServerError, "database error")
				}
			} else {
				now := time.Now()
				existingUser.LastLogin = &now
				existingUser.EmailVerified = emailVerified

				if err := db.Save(&existingUser).Error; err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user")
				}

				c.Set("current_user", &existingUser)
				c.Set("user_id", existingUser.ID)
			}

			return next(c)
		}
	}
}