package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Zolet-hash/healthcare-mcp-gateway/internal/auth"
	"github.com/Zolet-hash/healthcare-mcp-gateway/internal/healthcare"
	"github.com/Zolet-hash/healthcare-mcp-gateway/internal/httpserver"
)

func main() {
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtIssuer := os.Getenv("JWT_ISSUER")
	jwtAudience := os.Getenv("JWT_AUDIENCE")

	authConfig := auth.Config{
		Secret:   []byte(jwtSecret),
		Issuer:   jwtIssuer,
		Audience: jwtAudience,
	}

	if err := authConfig.Validate(); err != nil {
		log.Fatalf("invalid authentication configuration: %v", err)
	}

	patients := healthcare.NewPatientService()
	router := httpserver.NewRouter(patients, authConfig)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("healthcare MCP gateway listening on :%s", port)

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
