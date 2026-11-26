package user

import (
	"errors"
	"net/http"

	"github.com/Massad/gin-boilerplate/internal/modules/tenant"
	"github.com/Massad/gin-boilerplate/internal/platform/validate"
	"github.com/gin-gonic/gin"
)

// Handler serves the user login and register handlers (public routes: no
// TokenAuth).
type Handler struct {
	users *Service
}

// NewHandler builds the user handlers on the injected service.
func NewHandler(users *Service) Handler {
	return Handler{users: users}
}

// Login User godoc
// @Summary Login User example
// @Schemes
// @Description Login User example
// @Tags User
// @Accept json
// @Produce json
// @Param article body user.LoginForm true "User"
// @Success 	 200  {object}  user.UserLoginResponse
// @Failure      406  {object}  models.MessageResponse
// @Router /user/login [post]
func (h Handler) Login(c *gin.Context) {
	var loginForm LoginForm

	if validationErr := c.ShouldBindJSON(&loginForm); validationErr != nil {
		message := validate.Translate(validationErr, LoginMessages)
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": message})
		return
	}

	user, tenant, token, err := h.users.Login(loginForm)
	if err != nil {
		if errors.Is(err, ErrTenantSuspended) {
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
// @Param article body user.RegisterForm true "User"
// @Success 	 200  {object}  user.UserLoginResponse
// @Failure      406  {object}  models.MessageResponse
// @Router /user/register [post]
func (h Handler) Register(c *gin.Context) {
	var registerForm RegisterForm

	if validationErr := c.ShouldBindJSON(&registerForm); validationErr != nil {
		message := validate.Translate(validationErr, RegisterMessages)
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": message})
		return
	}

	user, createdTenant, err := h.users.Register(registerForm)
	if err != nil {
		if errors.Is(err, tenant.ErrInvalidSlug) || errors.Is(err, tenant.ErrSlugUnavailable) {
			c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": err.Error()})
			return
		}
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": err.Error()})
		return
	}

	payload := TenantPayload{ID: createdTenant.ID.String(), Slug: createdTenant.Slug, Name: createdTenant.Name}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully registered", "user": user, "tenant": payload})
}
