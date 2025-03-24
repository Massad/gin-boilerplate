package controllers

import (
	"github.com/Massad/gin-boilerplate/forms"
	"github.com/Massad/gin-boilerplate/models"

	"net/http"

	"github.com/gin-gonic/gin"
)

// UserController ...
type UserController struct{}

var userModel = new(models.UserModel)
var userForm = new(forms.UserForm)

// getUserID ...
func getUserID(c *gin.Context) (userID int64) {
	//MustGet returns the value for the given key if it exists, otherwise it panics.
	return c.MustGet("userID").(int64)
}

// Login User godoc
// @Summary Login User example
// @Schemes
// @Description Login User example
// @Tags User
// @Accept json
// @Produce json
// @Param article body forms.LoginForm true "User"
// @Success 	 200  {object}  models.UserLoginResponse
// @Failure      406  {object}  models.MessageResponse
// @Router /user/login [post]
func (ctrl UserController) Login(c *gin.Context) {
	var loginForm forms.LoginForm

	if validationErr := c.ShouldBindJSON(&loginForm); validationErr != nil {
		message := userForm.Login(validationErr)
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": message})
		return
	}

	user, token, err := userModel.Login(loginForm)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": "Invalid login details"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged in", "user": user, "token": token})
}

// Register User godoc
// @Summary Register User example
// @Schemes
// @Description Register User example
// @Tags User
// @Accept json
// @Produce json
// @Param article body forms.RegisterForm true "User"
// @Success 	 200  {object}  models.UserLoginResponse
// @Failure      406  {object}  models.MessageResponse
// @Router /user/register [post]
func (ctrl UserController) Register(c *gin.Context) {
	var registerForm forms.RegisterForm

	if validationErr := c.ShouldBindJSON(&registerForm); validationErr != nil {
		message := userForm.Register(validationErr)
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": message})
		return
	}

	user, err := userModel.Register(registerForm)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully registered", "user": user})
}

// Logout User godoc
// @Summary Logout User example
// @Schemes
// @Description Logout User example
// @Tags User
// @Accept json
// @Produce json
// @Success 	 200  {object}  models.MessageResponse
// @Failure      406  {object}  models.MessageResponse
// @Router /user/logout [GET]
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
