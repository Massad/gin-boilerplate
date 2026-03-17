package middleware

import (
	"net/http"

	"github.com/Massad/gin-boilerplate/models"
	"github.com/gin-gonic/gin"
)

var authModel = new(models.AuthModel)

// TokenAuth validates the JWT access token and sets userID in the context.
func TokenAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenAuth, err := authModel.ExtractTokenMetadata(c.Request)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Please login first"})
			return
		}

		userID, err := authModel.FetchAuth(tokenAuth)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Please login first"})
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}
