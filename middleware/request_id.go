package middleware

import (
	"github.com/gin-gonic/gin"
	uuid "github.com/google/uuid"
)

// RequestID attaches a unique ID to each request via the X-Request-Id header.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid := uuid.New()
		c.Writer.Header().Set("X-Request-Id", uuid.String())
		c.Next()
	}
}
