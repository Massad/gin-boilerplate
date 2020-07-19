package controllers

import (
	"github.com/Massad/gin-boilerplate/forms"
	"github.com/Massad/gin-boilerplate/models"

	"net/http"

	"github.com/gin-gonic/gin"
)

//UserController ...
type UserController struct{}

var userModel = new(models.UserModel)

//getUserID ...
func getUserID(c *gin.Context) (userID int64) {
	//MustGet returns the value for the given key if it exists, otherwise it panics.
	return c.MustGet("userID").(int64)
}

//Login ...
func (ctrl UserController) Login(c *gin.Context) {
	var loginForm forms.LoginForm

	if c.ShouldBindJSON(&loginForm) != nil {
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": "Invalid form"})
		return
	}

	user, token, err := userModel.Login(loginForm)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "User signed in", "user": user, "token": token})
	} else {
		c.JSON(http.StatusNotAcceptable, gin.H{"message": "Invalid login details", "error": err.Error()})
	}

}

//Register ...
func (ctrl UserController) Register(c *gin.Context) {
	var registerForm forms.RegisterForm

	if c.ShouldBindJSON(&registerForm) != nil {
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": "Invalid form"})
		return
	}

	user, err := userModel.Register(registerForm)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": err.Error()})
		return
	}

	if user.ID > 0 {
		c.JSON(http.StatusOK, gin.H{"message": "Successfully registered", "user": user})
	} else {
		c.JSON(http.StatusNotAcceptable, gin.H{"message": "Could not register this user", "error": err.Error()})
	}

}

//Logout ...
func (ctrl UserController) Logout(c *gin.Context) {

	au, err := authModel.ExtractTokenMetadata(c.Request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "User not logged in"})
		return
	}

	deleted, delErr := authModel.DeleteAuth(au.AccessUUID)
	if delErr != nil || deleted == 0 { //if any goes wrong
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}
