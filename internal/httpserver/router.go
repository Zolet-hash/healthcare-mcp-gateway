package httpserver

import (
	"net/http"

	"github.com/Zolet-hash/healthcare-mcp-gateway/internal/auth"
	"github.com/Zolet-hash/healthcare-mcp-gateway/internal/healthcare"
	"github.com/Zolet-hash/healthcare-mcp-gateway/internal/mcpserver"
	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func NewRouter(
	patients *healthcare.PatientService,
	authConfig auth.Config,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	handler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server {
			identity, ok := auth.IdentityFromRequest(r)
			if !ok {
				// The middleware should prevent this path.
				// An empty tenant must never grant access.
				return mcpserver.New(patients, "")
			}

			return mcpserver.New(patients, identity.TenantID)
		},
		&mcp.StreamableHTTPOptions{
			JSONResponse: true,
			Stateless:    true,
		},
	)

	// Protect every MCP HTTP request with bearer-token authentication.
	router.Any("/mcp", auth.Middleware(authConfig), gin.WrapH(handler))

	return router
}
