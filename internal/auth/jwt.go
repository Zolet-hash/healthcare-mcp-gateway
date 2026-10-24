
package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	Secret   []byte
	Issuer   string
	Audience string
}

type Identity struct {
	UserID   string
	TenantID string
}

type Claims struct {
	TenantID string `json:"tenant_id"`
	jwt.RegisteredClaims
}

func (c Config) Validate() error {
	if len(c.Secret) < 32 {
		return errors.New("JWT secret must be at least 32 bytes")
	}
	if strings.TrimSpace(c.Issuer) == "" {
		return errors.New("JWT issuer is required")
	}
	if strings.TrimSpace(c.Audience) == "" {
		return errors.New("JWT audience is required")
	}
	return nil
}

func VerifyToken(tokenString string, cfg Config) (Identity, error) {
	var identity Identity

	if err := cfg.Validate(); err != nil {
		return identity, fmt.Errorf("invalid auth configuration: %w", err)
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected JWT signing algorithm")
			}
			return cfg.Secret, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer(cfg.Issuer),
		jwt.WithAudience(cfg.Audience),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(30*time.Second),
	)
	if err != nil {
		return identity, fmt.Errorf("invalid bearer token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return identity, errors.New("invalid token claims")
	}

	if strings.TrimSpace(claims.Subject) == "" {
		return identity, errors.New("token subject is required")
	}
	if strings.TrimSpace(claims.TenantID) == "" {
		return identity, errors.New("token tenant_id is required")
	}

	identity.UserID = claims.Subject
	identity.TenantID = claims.TenantID

	return identity, nil
}
