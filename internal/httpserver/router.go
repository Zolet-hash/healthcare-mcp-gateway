package httpserver

import (
	"net/http"

	"github.com/Zolet-hash/healthcare-mcp-gateway/internal/healthcare"
	"github.com/Zolet-hash/healthcare-mcp-gateway/internal/mcpserver"
	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func NewRouter(patients *healthcare.PatientService) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	handler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server {
			// Development-only tenant identity.
			// Replace with verified authentication in production.
			return mcpserver.New(
				patients,
				r.Header.Get("X-Tenant-ID"),
			)
		},
		&mcp.StreamableHTTPOptions{
			JSONResponse: true,
			Stateless:    true,
		},
	)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.Any("/mcp", func(c *gin.Context) {
		// Never trust a caller-supplied tenant identity in production.
		if c.GetHeader("X-Tenant-ID") == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing development tenant identity",
			})
			return
		}

		handler.ServeHTTP(c.Writer, c.Request)
	})

	return router
}
