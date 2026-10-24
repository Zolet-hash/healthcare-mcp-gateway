package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Zolet-hash/healthcare-mcp-gateway/internal/auth"
	"github.com/Zolet-hash/healthcare-mcp-gateway/internal/healthcare"
	"github.com/gin-gonic/gin"
)

func TestHealthEndpointDoesNotRequireAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := auth.Config{
		Secret:   []byte("router-test-secret-at-least-32-bytes-long"),
		Issuer:   "router-test",
		Audience: "healthcare-mcp-gateway",
	}

	router := NewRouter(healthcare.NewPatientService(), cfg)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestMCPRejectsInvalidBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := auth.Config{
		Secret:   []byte("router-test-secret-at-least-32-bytes-long"),
		Issuer:   "router-test",
		Audience: "healthcare-mcp-gateway",
	}

	router := NewRouter(healthcare.NewPatientService(), cfg)
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer definitely-not-a-valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusUnauthorized,
		)
	}
}
