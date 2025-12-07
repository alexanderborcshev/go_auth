package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RoleMiddleware ensures the user has one of the required roles.
func RoleMiddleware(requiredRoles ...string) gin.HandlerFunc {
	allowed := map[string]struct{}{}
	for _, r := range requiredRoles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		roleVal, exists := c.Get(CtxUserRoleKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		roleStr, _ := roleVal.(string)
		if _, ok := allowed[roleStr]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient role"})
			return
		}
		c.Next()
	}
}
