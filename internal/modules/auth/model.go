package auth

import "github.com/google/uuid"

// AuthResponse is the documented payload of a refresh response.
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// TokenDetails ...
type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUUID   string
	RefreshUUID  string
	AtExpires    int64
	RtExpires    int64
}

// AdminTokenDetails carries the single, long-lived superadmin access token.
type AdminTokenDetails struct {
	AccessToken string
	AccessUUID  string
	AtExpires   int64
}

// AccessDetails carries the tenant-scoped session identity extracted from a
// access token: an admin token is rejected here because it lacks the claims.
type AccessDetails struct {
	AccessUUID string
	UserID     uuid.UUID
	TenantID   uuid.UUID
	Role       string
}

// AdminAccessDetails carries the platform session identity of an admin token.
type AdminAccessDetails struct {
	AccessUUID string
	AdminID    uuid.UUID
}

// Token is the session token pair of a tenant session. RefreshToken is
// omitted when empty: the unified login reuses this model for its superadmin
// branch, which carries an access token only (no refresh half).
type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// AdminToken is the token payload of a superadmin session (no refresh).
type AdminToken struct {
	AccessToken string `json:"access_token"`
}

// Secrets carries the three independent signing secrets of the session
// tokens. They are injected at construction by the composition root: the
// service never reads the environment (tests set them with t.Setenv).
type Secrets struct {
	Access      string // ACCESS_SECRET signs the tenant access tokens
	Refresh     string // REFRESH_SECRET signs the tenant refresh tokens
	AdminAccess string // ADMIN_ACCESS_SECRET signs the superadmin access token
}
