package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Permissions of the minimal RBAC layer: owner-only routes opt into them later.
const (
	PermManageMembers = "members:manage"
	PermManageTenant  = "tenant:manage"
)

// rolePermissions maps a user role to the permissions it holds.
var rolePermissions = map[string][]string{
	"owner":  {PermManageMembers, PermManageTenant},
	"member": {},
}

// hasPermission reports whether a role holds the given permission.
func hasPermission(role, permission string) bool {
	for _, p := range rolePermissions[role] {
		if p == permission {
			return true
		}
	}
	return false
}

// TokenAuth validates the JWT access token, resolves the session in Redis and
// exposes userID, tenantID and role to the handler chain (context keys
// "userID", "tenantID", "role"). An admin token is rejected: it is signed with
// a different secret and has no tenant claims.
func TokenAuth(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		details, err := svc.ExtractTokenMetadata(c.Request)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Please login first"})
			return
		}

		userID, err := svc.FetchAuth(details)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Please login first"})
			return
		}

		c.Set("userID", userID.String())
		c.Set("tenantID", details.TenantID.String())
		c.Set("role", details.Role)
		c.Next()
	}
}

// RequirePermission blocks authenticated users whose role lacks the permission
// (403: authenticated but not allowed).
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !hasPermission(c.GetString("role"), permission) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Insufficient permissions"})
			return
		}
		c.Next()
	}
}

// AdminAuth validates the superadmin access token, signed with the dedicated
// secret and stored under the admin: Redis prefix: a tenant token cannot pass.
// The resolved identity is exposed under the context key "adminID".
func AdminAuth(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		details, err := svc.ExtractAdminTokenMetadata(c.Request)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Admin authentication required"})
			return
		}

		adminID, err := svc.FetchAdmin(details)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Admin authentication required"})
			return
		}

		c.Set("adminID", adminID.String())
		c.Next()
	}
}
