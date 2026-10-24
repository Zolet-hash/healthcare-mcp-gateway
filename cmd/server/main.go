package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Zolet-hash/healthcare-mcp-gateway/internal/healthcare"
	"github.com/Zolet-hash/healthcare-mcp-gateway/internal/httpserver"
)

func main() {
	patients := healthcare.NewPatientService()
	router := httpserver.NewRouter(patients)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("Healthcare MCP gateway listening on :%s", port)
	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
