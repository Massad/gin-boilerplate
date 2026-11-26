package user

import (
	auth "github.com/Massad/gin-boilerplate/internal/modules/auth"
	"github.com/google/uuid"
)

// UserLoginResponse ...
type UserLoginResponse struct {
	User    User          `json:"user"`
	Tenant  TenantPayload `json:"tenant"`
	Token   auth.Token    `json:"token"`
	Message string        `json:"message"`
}

// TenantPayload is the tenant subset embedded in user auth responses. The
// payload shape is carried by each consumer (design Phase 5): the port
// UserAuthenticator defines its own in the auth module; no cross-module type
// import.
type TenantPayload struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// User ...
type User struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	// Password is the bcrypt hash; never serialized to clients.
	Password string    `json:"-"`
	Name     string    `json:"name"`
	TenantID uuid.UUID `json:"tenant_id"`
	Role     string    `json:"role"`
}

// UserWithTenantStatus bundles the refresh-time re-read of a user with the
// freshness data the refresh flow needs.
type UserWithTenantStatus struct {
	User       User
	Tenant     TenantPayload
	TenantSlug string
	Status     int16
}
