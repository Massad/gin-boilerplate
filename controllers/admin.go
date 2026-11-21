package controllers

import (
	"net/http"

	"github.com/Massad/gin-boilerplate/forms"
	"github.com/Massad/gin-boilerplate/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminController ...
type AdminController struct{}

var adminModel = new(models.AdminModel)

// Login Admin godoc
// @Summary Login Admin example
// @Schemes
// @Description Login Admin example
// @Tags Admin
// @Accept json
// @Produce json
// @Param auth body forms.LoginForm true "Admin"
// @Success 	 200  {object}  models.AuthResponse
// @Failure      406  {object}  models.MessageResponse
// @Router /admin/login [post]
func (ctrl AdminController) Login(c *gin.Context) {
	var loginForm forms.LoginForm

	if validationErr := c.ShouldBindJSON(&loginForm); validationErr != nil {
		message := forms.Translate(validationErr, forms.LoginMessages)
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": message})
		return
	}

	adminID, err := adminModel.Login(loginForm.Email, loginForm.Password)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": "Invalid admin login details"})
		return
	}

	tokenDetails, err := authModel.CreateAdminToken(adminID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Invalid admin login details"})
		return
	}
	if err := authModel.CreateAdminAuth(adminID, tokenDetails); err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Invalid admin login details"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged in", "token": models.AdminToken{AccessToken: tokenDetails.AccessToken}})
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
func (ctrl AdminController) Tenants(c *gin.Context) {
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
func (ctrl AdminController) SetTenantStatus(c *gin.Context) {
	adminID, _ := uuid.Parse(c.GetString("adminID"))

	id := c.Param("id")

	tenantID, err := uuid.Parse(id)
	if err != nil || tenantID == uuid.Nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"Message": "Invalid parameter"})
		return
	}

	var form forms.TenantStatusForm

	if validationErr := c.ShouldBindJSON(&form); validationErr != nil {
		message := forms.Translate(validationErr, forms.AdminStatusMessages)
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

// Logout Admin godoc
// @Summary Logout Admin example
// @Schemes
// @Description Logout Admin example
// @Tags Admin
// @Success 	 200  {object}  models.MessageResponse
// @Failure      406  {object}  models.MessageResponse
// @Router /admin/logout [get]
func (ctrl AdminController) Logout(c *gin.Context) {
	details, err := authModel.ExtractAdminTokenMetadata(c.Request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Admin not logged in"})
		return
	}

	deleted, delErr := authModel.DeleteAdminAuth(details.AccessUUID)
	if delErr != nil || deleted == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}
