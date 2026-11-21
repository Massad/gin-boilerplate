package controllers

import (
	"net/http"

	"github.com/Massad/gin-boilerplate/models"
	"github.com/gin-gonic/gin"
)

// PublicController ...
type PublicController struct{}

var publicTenantModel = new(models.TenantModel)

// One godoc
// @Summary Public tenant lookup by slug
// @Schemes
// @Description Expose the public directory data of an active tenant
// @Tags Public
// @Produce json
// @Success 	 200  {object}  models.MessageResponse
// @Failure      404  {object}  models.MessageResponse
// @Router /public/{slug} [get]
func (ctrl PublicController) One(c *gin.Context) {
	slug := c.Param("slug")

	// The slug follows the same pattern as the registration: an invalid path
	// is simply an unknown tenant.
	if !models.ValidSlug(slug) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Tenant not found"})
		return
	}

	tenant, err := publicTenantModel.OneBySlug(slug)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Tenant not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tenant": tenant})
}
