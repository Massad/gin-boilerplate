// Package httpx groups the shared HTTP responders, the pgx-to-HTTP error
// mapping and the route parameter parsing used by the module handlers.
package httpx

import (
	"github.com/gin-gonic/gin"
)

// AbortMessage aborts the in-flight request with the standard JSON error
// shape {"message": ...} and the given status code.
func AbortMessage(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"message": message})
}

// JSONMessage responds with the standard JSON shape {"message": ...} and the
// given status code.
func JSONMessage(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"message": message})
}
