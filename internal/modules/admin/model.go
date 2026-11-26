package admin

import "github.com/google/uuid"

// AdminIdentity is the superadmin identity verified at login time. It is
// embedded in the unified login response.
type AdminIdentity struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Name  string    `json:"name"`
}
