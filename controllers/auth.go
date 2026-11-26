package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/Massad/gin-boilerplate/internal/modules/admin"
	auth "github.com/Massad/gin-boilerplate/internal/modules/auth"
	user "github.com/Massad/gin-boilerplate/internal/modules/user"

	"github.com/Massad/gin-boilerplate/forms"
	"github.com/Massad/gin-boilerplate/internal/platform/validate"
	"github.com/Massad/gin-boilerplate/models"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// AuthController serves the unified login and the refresh flow during the
// transitory migration window: the handlers move into the auth module in
// Phase 7.
type AuthController struct {
	auth *auth.Service
	// users is the user module service injected at the composition root:
	// the unified login runs its Login and the refresh flow re-reads the
	// user through OneWithTenant.
	users *user.Service
	// admins is the admin module service injected at the composition root:
	// the unified login runs its superadmin fallback through it.
	admins *admin.Service
}

// NewAuthController builds the controller on the injected auth, user and
// admin services.
func NewAuthController(svc *auth.Service, users *user.Service, admins *admin.Service) *AuthController {
	return &AuthController{auth: svc, users: users, admins: admins}
}

// Login Unified godoc
// @Summary Unified login example
// @Schemes
// @Description Unified login: tenant user first, superadmin fallback
// @Description The response is discriminated by kind: kind=tenant_user carries user, tenant and an access/refresh token pair; kind=superadmin carries admin and an access token only (no refresh_token).
// @Tags Auth
// @Accept json
// @Produce json
// @Param auth body user.LoginForm true "Auth"
// @Success 	 200  {object}  models.UnifiedLoginResponse
// @Failure      403  {object}  models.MessageResponse
// @Failure      406  {object}  models.MessageResponse
// @Router /login [post]
func (ctl AuthController) Login(c *gin.Context) {
	var loginForm user.LoginForm

	if validationErr := c.ShouldBindJSON(&loginForm); validationErr != nil {
		message := validate.Translate(validationErr, user.LoginMessages)
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": message})
		return
	}

	result, err := models.UnifiedLogin(ctl.auth, ctl.users, ctl.admins, loginForm)
	if err != nil {
		if errors.Is(err, user.ErrTenantSuspended) {
			// A suspended tenant stays observable: the 403 is not masked.
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Tenant suspended"})
			return
		}
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": "Invalid login details"})
		return
	}

	if result.Kind == models.LoginKindSuperadmin {
		// The superadmin session has no refresh half: the token model omits
		// the empty refresh_token key.
		c.JSON(http.StatusOK, gin.H{
			"kind":  result.Kind,
			"admin": result.Admin,
			"token": auth.Token{AccessToken: result.AccessToken},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"kind":   result.Kind,
		"user":   result.User,
		"tenant": result.Tenant,
		"token":  result.Token,
	})
}

// Refresh Token godoc
// @Summary Refresh Token example
// @Schemes
// @Description Refresh Token example
// @Tags Auth
// @Accept json
// @Produce json
// @Param auth body forms.Token true "Auth"
// @Success 	 200  {object}  auth.AuthResponse
// @Failure      406  {object}  models.MessageResponse
// @Router /token/refresh [POST]
func (ctl AuthController) Refresh(c *gin.Context) {
	var tokenForm forms.Token

	if c.ShouldBindJSON(&tokenForm) != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"message": "Invalid form", "form": tokenForm})
		c.Abort()
		return
	}

	//verify the token
	token, err := jwt.Parse(tokenForm.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("REFRESH_SECRET")), nil
	})
	//if there is an error, the token must have expired
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid authorization, please login again"})
		return
	}
	//is token valid?
	claims, ok := token.Claims.(jwt.MapClaims) //the token claims should conform to MapClaims
	if !ok || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid authorization, please login again"})
		return
	}
	refreshUUID, ok := claims["refresh_uuid"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid authorization, please login again"})
		return
	}
	userIDClaim, ok := claims["user_id"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid authorization, please login again"})
		return
	}
	userID, err := uuid.Parse(userIDClaim)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid authorization, please login again"})
		return
	}
	// Verify ownership and softly consume the previous refresh token: 1
	// claims the first consumption (the record is marked delivered and a
	// failing flow must release it), 2 claims the single recovery of a
	// not-delivered record, 0 refuses the token.
	consumed, consumeErr := ctl.auth.ConsumeRefreshAuth(refreshUUID, userID)
	if consumeErr != nil || consumed == 0 { //if any goes wrong
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid authorization, please login again"})
		return
	}

	// Re-read the user through the platform pool: the role is never read back
	// from the claims, the user must still exist and its tenant stay active.
	row, readErr := ctl.users.OneWithTenant(userID)
	if readErr != nil {
		// The consumption delivered nothing: release the record so the
		// replayed token stays recoverable during the grace.
		if consumed == 1 {
			_, _ = ctl.auth.ReleaseRefresh(refreshUUID, userID)
		}
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid authorization, please login again"})
		return
	}
	if row.Status != 1 {
		if consumed == 1 {
			_, _ = ctl.auth.ReleaseRefresh(refreshUUID, userID)
		}
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid authorization, please login again"})
		return
	}

	//Create new pairs of refresh and access tokens
	ts, createErr := ctl.auth.CreateToken(row.User.ID, row.User.TenantID, row.User.Role)
	if createErr != nil {
		if consumed == 1 {
			_, _ = ctl.auth.ReleaseRefresh(refreshUUID, userID)
		}
		c.JSON(http.StatusForbidden, gin.H{"message": "Invalid authorization, please login again"})
		return
	}
	//save the tokens metadata to redis
	saveErr := ctl.auth.CreateAuth(row.User.ID, row.User.TenantID, row.User.Role, ts)
	if saveErr != nil {
		if consumed == 1 {
			_, _ = ctl.auth.ReleaseRefresh(refreshUUID, userID)
		}
		c.JSON(http.StatusForbidden, gin.H{"message": "Invalid authorization, please login again"})
		return
	}
	tokens := map[string]string{
		"access_token":  ts.AccessToken,
		"refresh_token": ts.RefreshToken,
	}
	c.JSON(http.StatusOK, tokens)
}
