package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler serves the module HTTP surface of Phase 3: the two logout handlers.
// The login handlers stay in the root controllers until Phases 6 and 7
// migrate them into the user/admin/auth modules.
type Handler struct {
	svc *Service
}

// NewHandler builds the auth handlers on the injected service.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
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
func (h *Handler) UserLogout(c *gin.Context) {
	au, err := h.svc.ExtractTokenMetadata(c.Request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "User not logged in"})
		return
	}

	deleted, delErr := h.svc.DeleteAuth(au.AccessUUID, au.UserID)
	if delErr != nil || deleted == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

// Logout Admin godoc
// @Summary Logout Admin example
// @Schemes
// @Description Logout Admin example
// @Tags Admin
// @Success 	 200  {object}  models.MessageResponse
// @Failure      406  {object}  models.MessageResponse
// @Router /admin/logout [get]
func (h *Handler) AdminLogout(c *gin.Context) {
	details, err := h.svc.ExtractAdminTokenMetadata(c.Request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Admin not logged in"})
		return
	}

	deleted, delErr := h.svc.DeleteAdminAuth(details.AccessUUID)
	if delErr != nil || deleted == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}
