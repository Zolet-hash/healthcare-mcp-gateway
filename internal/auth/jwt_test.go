package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func testConfig() Config {
	return Config{
		Secret:   []byte("local-test-secret-at-least-32-bytes-long"),
		Issuer:   "healthcare-test-issuer",
		Audience: "healthcare-mcp-gateway",
	}
}

func signedToken(t *testing.T, cfg Config, claims Claims) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(cfg.Secret)
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

func TestVerifyToken(t *testing.T) {
	cfg := testConfig()
	now := time.Now()

	raw := signedToken(t, cfg, Claims{
		TenantID: "tenant-demo",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			Issuer:    cfg.Issuer,
			Audience:  jwt.ClaimStrings{cfg.Audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	})

	got, err := VerifyToken(raw, cfg)
	if err != nil {
		t.Fatalf("VerifyToken() error = %v", err)
	}

	if got.UserID != "user-123" {
		t.Errorf("UserID = %q, want %q", got.UserID, "user-123")
	}
	if got.TenantID != "tenant-demo" {
		t.Errorf("TenantID = %q, want %q", got.TenantID, "tenant-demo")
	}
}

func TestVerifyTokenRejectsWrongSecret(t *testing.T) {
	cfg := testConfig()
	now := time.Now()

	raw := signedToken(t, cfg, Claims{
		TenantID: "tenant-demo",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			Issuer:    cfg.Issuer,
			Audience:  jwt.ClaimStrings{cfg.Audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
	})

	wrongConfig := cfg
	wrongConfig.Secret = []byte("another-test-secret-at-least-32-bytes")

	if _, err := VerifyToken(raw, wrongConfig); err == nil {
		t.Fatal("expected invalid signature to be rejected")
	}
}

func TestVerifyTokenRejectsExpiredToken(t *testing.T) {
	cfg := testConfig()

	raw := signedToken(t, cfg, Claims{
		TenantID: "tenant-demo",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			Issuer:    cfg.Issuer,
			Audience:  jwt.ClaimStrings{cfg.Audience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	})

	if _, err := VerifyToken(raw, cfg); err == nil {
		t.Fatal("expected expired token to be rejected")
	}
}
