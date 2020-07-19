package controllers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/Massad/gin-boilerplate/forms"
	"github.com/Massad/gin-boilerplate/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

//AuthController ...
type AuthController struct{}

var authModel = new(models.AuthModel)

//TokenValid ...
func (ctl AuthController) TokenValid(c *gin.Context) {

	tokenAuth, err := authModel.ExtractTokenMetadata(c.Request)
	if err != nil {
		//Token either expired or not valid
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Please login first"})
		return
	}

	userID, err := authModel.FetchAuth(tokenAuth)
	if err != nil {
		//Token does not exists in Redis (User logged out or expired)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Please login first"})
		return
	}

	//To be called from GetUserID()
	c.Set("userID", userID)
}

//Refresh ...
func (ctl AuthController) Refresh(c *gin.Context) {
	var tokenForm forms.Token

	if c.ShouldBindJSON(&tokenForm) != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"message": "Invalid form", "form": tokenForm})
		c.Abort()
		return
	}

	//verify the token
	token, err := jwt.Parse(tokenForm.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("REFRESH_SECRET")), nil
	})
	//if there is an error, the token must have expired
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid authorization, please login again"})
		return
	}
	//is token valid?
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid authorization, please login again"})
		return
	}
	//Since token is valid, get the uuid:
	claims, ok := token.Claims.(jwt.MapClaims) //the token claims should conform to MapClaims
	if ok && token.Valid {
		refreshUUID, ok := claims["refresh_uuid"].(string) //convert the interface to string
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid authorization, please login again"})
			return
		}
		userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid authorization, please login again"})
			return
		}
		//Delete the previous Refresh Token
		deleted, delErr := authModel.DeleteAuth(refreshUUID)
		if delErr != nil || deleted == 0 { //if any goes wrong
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid authorization, please login again"})
			return
		}

		//Create new pairs of refresh and access tokens
		ts, createErr := authModel.CreateToken(userID)
		if createErr != nil {
			c.JSON(http.StatusForbidden, gin.H{"message": "Invalid authorization, please login again"})
			return
		}
		//save the tokens metadata to redis
		saveErr := authModel.CreateAuth(userID, ts)
		if saveErr != nil {
			c.JSON(http.StatusForbidden, gin.H{"message": "Invalid authorization, please login again"})
			return
		}
		tokens := map[string]string{
			"access_token":  ts.AccessToken,
			"refresh_token": ts.RefreshToken,
		}
		c.JSON(http.StatusOK, tokens)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid authorization, please login again"})
	}
}
