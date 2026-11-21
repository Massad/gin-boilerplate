package controllers

import (
	"errors"
	"net/http"

	"github.com/Massad/gin-boilerplate/forms"
	"github.com/Massad/gin-boilerplate/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserController ...
type UserController struct{}

var userModel = new(models.UserModel)

// getUserID reads the authenticated user id set by TokenAuth.
func getUserID(c *gin.Context) (userID uuid.UUID) {
	userID, _ = uuid.Parse(c.GetString("userID"))
	return userID
}

// getTenantID reads the authenticated tenant id set by TokenAuth.
func getTenantID(c *gin.Context) (tenantID uuid.UUID) {
	tenantID, _ = uuid.Parse(c.GetString("tenantID"))
	return tenantID
}

// getRole reads the authenticated role set by TokenAuth.
func getRole(c *gin.Context) (role string) {
	return c.GetString("role")
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
		message := forms.Translate(validationErr, forms.LoginMessages)
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": message})
		return
	}

	user, tenant, token, err := userModel.Login(loginForm)
	if err != nil {
		if errors.Is(err, models.ErrTenantSuspended) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Tenant suspended"})
			return
		}
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": "Invalid login details"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged in", "user": user, "tenant": tenant, "token": token})
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
		message := forms.Translate(validationErr, forms.RegisterMessages)
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": message})
		return
	}

	user, tenant, err := userModel.Register(registerForm)
	if err != nil {
		if errors.Is(err, models.ErrInvalidSlug) || errors.Is(err, models.ErrSlugUnavailable) {
			c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": err.Error()})
			return
		}
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": err.Error()})
		return
	}

	payload := models.TenantPayload{ID: tenant.ID.String(), Slug: tenant.Slug, Name: tenant.Name}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully registered", "user": user, "tenant": payload})
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

	deleted, delErr := authModel.DeleteAuth(au.AccessUUID, au.UserID)
	if delErr != nil || deleted == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}
