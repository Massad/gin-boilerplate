package httpx

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ParseUUIDParam parses the named route parameter as a UUID. Whether the
// parsed value is the zero UUID stays a caller decision.
func ParseUUIDParam(c *gin.Context, name string) (uuid.UUID, error) {
	return uuid.Parse(c.Param(name))
}
