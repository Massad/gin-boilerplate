package controllers

import (
	"errors"
	"net/http"

	"github.com/Massad/gin-boilerplate/forms"
	"github.com/Massad/gin-boilerplate/internal/platform/validate"
	"github.com/Massad/gin-boilerplate/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserController serves the user login and register handlers during the
// transitory migration window: they move into the user module in Phase 5.
type UserController struct {
	users models.UserModel
}

var userModel = new(models.UserModel)

// NewUserController builds the controller on the injected user model (which
// consumes the *auth.Service at the composition root).
func NewUserController(users models.UserModel) *UserController {
	return &UserController{users: users}
}

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
func (ctl UserController) Login(c *gin.Context) {
	var loginForm forms.LoginForm

	if validationErr := c.ShouldBindJSON(&loginForm); validationErr != nil {
		message := validate.Translate(validationErr, forms.LoginMessages)
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": message})
		return
	}

	user, tenant, token, err := ctl.users.Login(loginForm)
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
func (ctl UserController) Register(c *gin.Context) {
	var registerForm forms.RegisterForm

	if validationErr := c.ShouldBindJSON(&registerForm); validationErr != nil {
		message := validate.Translate(validationErr, forms.RegisterMessages)
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": message})
		return
	}

	user, tenant, err := ctl.users.Register(registerForm)
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
