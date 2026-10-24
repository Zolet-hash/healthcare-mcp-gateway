
package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type identityContextKey struct{}

const IdentityContextKey = "authenticated_identity"

func Middleware(cfg Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.Fields(header)

		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing or invalid bearer token",
			})
			return
		}

		identity, err := VerifyToken(parts[1], cfg)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired bearer token",
			})
			return
		}

		// Store the identity in Gin's context.
		c.Set(IdentityContextKey, identity)

		// Also store it in the request context for net/http handlers,
		// including the MCP SDK's Streamable HTTP handler.
		ctx := context.WithValue(c.Request.Context(), identityContextKey{}, identity)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func IdentityFromContext(c *gin.Context) (Identity, bool) {
	value, exists := c.Get(IdentityContextKey)
	if !exists {
		return Identity{}, false
	}

	identity, ok := value.(Identity)
	return identity, ok
}

func IdentityFromRequest(r *http.Request) (Identity, bool) {
	identity, ok := r.Context().Value(identityContextKey{}).(Identity)
	return identity, ok
}
