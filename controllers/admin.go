package controllers

import (
	"net/http"

	auth "github.com/Massad/gin-boilerplate/internal/modules/auth"

	"github.com/Massad/gin-boilerplate/forms"
	"github.com/Massad/gin-boilerplate/internal/platform/validate"
	"github.com/Massad/gin-boilerplate/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminController serves the backoffice handlers during the transitory
// migration window: the model moves into the admin module in Phase 6.
type AdminController struct {
	auth *auth.Service
}

var adminModel = new(models.AdminModel)

// NewAdminController builds the controller on the injected session service.
func NewAdminController(svc *auth.Service) *AdminController {
	return &AdminController{auth: svc}
}

// Login Admin godoc
// @Summary Login Admin example
// @Schemes
// @Description Login Admin example
// @Tags Admin
// @Accept json
// @Produce json
// @Param auth body forms.LoginForm true "Admin"
// @Success 	 200  {object}  auth.AuthResponse
// @Failure      406  {object}  models.MessageResponse
// @Router /admin/login [post]
func (ctl AdminController) Login(c *gin.Context) {
	var loginForm forms.LoginForm

	if validationErr := c.ShouldBindJSON(&loginForm); validationErr != nil {
		message := validate.Translate(validationErr, forms.LoginMessages)
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": message})
		return
	}

	admin, err := adminModel.Login(loginForm.Email, loginForm.Password)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": "Invalid admin login details"})
		return
	}

	tokenDetails, err := ctl.auth.CreateAdminToken(admin.ID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Invalid admin login details"})
		return
	}
	if err := ctl.auth.CreateAdminAuth(admin.ID, tokenDetails); err != nil {
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
// @Success 	 200  {object}  models.AllTenantResponse
// @Failure      406  {object}  models.MessageResponse
// @Router /admin/tenants [get]
func (ctl AdminController) Tenants(c *gin.Context) {
	tenants, err := adminModel.AllTenants()
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
// @Param status body forms.TenantStatusForm true "Status"
// @Success 	 200  {object}  models.MessageResponse
// @Failure      406  {object}  models.MessageResponse
// @Router /admin/tenants/{id}/status [post]
func (ctl AdminController) SetTenantStatus(c *gin.Context) {
	adminID, _ := uuid.Parse(c.GetString("adminID"))

	id := c.Param("id")

	tenantID, err := uuid.Parse(id)
	if err != nil || tenantID == uuid.Nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"Message": "Invalid parameter"})
		return
	}

	var form forms.TenantStatusForm

	if validationErr := c.ShouldBindJSON(&form); validationErr != nil {
		message := validate.Translate(validationErr, forms.AdminStatusMessages)
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": message})
		return
	}

	if err := adminModel.SetTenantStatus(adminID, tenantID, *form.Status); err != nil {
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": "Tenant status could not be updated"})
		return
	}

	message := "Tenant activated"
	if *form.Status == 0 {
		message = "Tenant suspended"
	}

	c.JSON(http.StatusOK, gin.H{"message": message})
}
