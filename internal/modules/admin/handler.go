package admin

import (
	"net/http"

	auth "github.com/Massad/gin-boilerplate/internal/modules/auth"
	"github.com/Massad/gin-boilerplate/internal/platform/validate"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler serves the backoffice handlers (the admin login is public, the
// tenants/status routes sit behind AdminAuth). The token issuance flows
// through the TokenService port declared by the module: the implementation
// *auth.Service is injected at the composition root.
type Handler struct {
	admins *Service
	tokens TokenService
}

// NewHandler builds the admin handlers on the injected service and token
// service.
func NewHandler(admins *Service, tokens TokenService) Handler {
	return Handler{admins: admins, tokens: tokens}
}

// Login Admin godoc
// @Summary Login Admin example
// @Schemes
// @Description Login Admin example
// @Tags Admin
// @Accept json
// @Produce json
// @Param auth body admin.LoginForm true "Admin"
// @Success 	 200  {object}  auth.AdminToken
// @Failure      406  {object}  models.MessageResponse
// @Router /admin/login [post]
func (h Handler) Login(c *gin.Context) {
	var loginForm LoginForm

	if validationErr := c.ShouldBindJSON(&loginForm); validationErr != nil {
		message := validate.Translate(validationErr, LoginMessages)
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": message})
		return
	}

	admin, err := h.admins.Login(loginForm.Email, loginForm.Password)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": "Invalid admin login details"})
		return
	}

	tokenDetails, err := h.tokens.CreateAdminToken(admin.ID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Invalid admin login details"})
		return
	}
	if err := h.tokens.CreateAdminAuth(admin.ID, tokenDetails); err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Invalid admin login details"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged in", "token": auth.AdminToken{AccessToken: tokenDetails.AccessToken}})
}

// Tenants godoc
// @Summary List tenants (admin)
// @Schemes
// @Description List tenants for the platform backoffice
// @Tags Admin
// @Produce json
// @Success 	 200  {object}  tenant.AllTenantResponse
// @Failure      406  {object}  models.MessageResponse
// @Router /admin/tenants [get]
func (h Handler) Tenants(c *gin.Context) {
	tenants, err := h.admins.AllTenants()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": "Could not list tenants"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tenants": tenants})
}

// SetTenantStatus godoc
// @Summary Set tenant status (admin)
// @Schemes
// @Description Suspend or reactivate a tenant and log the audit line
// @Tags Admin
// @Accept json
// @Produce json
// @Param status body admin.TenantStatusForm true "Status"
// @Success 	 200  {object}  models.MessageResponse
// @Failure      406  {object}  models.MessageResponse
// @Router /admin/tenants/{id}/status [post]
func (h Handler) SetTenantStatus(c *gin.Context) {
	adminID, _ := uuid.Parse(c.GetString("adminID"))

	id := c.Param("id")

	tenantID, err := uuid.Parse(id)
	if err != nil || tenantID == uuid.Nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"Message": "Invalid parameter"})
		return
	}

	var form TenantStatusForm

	if validationErr := c.ShouldBindJSON(&form); validationErr != nil {
		message := validate.Translate(validationErr, AdminStatusMessages)
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": message})
		return
	}

	if err := h.admins.SetTenantStatus(adminID, tenantID, *form.Status); err != nil {
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": "Tenant status could not be updated"})
		return
	}

	message := "Tenant activated"
	if *form.Status == 0 {
		message = "Tenant suspended"
	}

	c.JSON(http.StatusOK, gin.H{"message": message})
}
