
package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

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

		c.Set(IdentityContextKey, identity)
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
