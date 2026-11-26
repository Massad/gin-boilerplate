package tenant

import (
	"time"

	"github.com/google/uuid"
)

// Tenant represents one organization sharing the platform database.
type Tenant struct {
	ID        uuid.UUID `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Status    int16     `json:"status"`
	CreatedAt int64     `json:"-"`
	UpdatedAt int64     `json:"-"`
}

// AllTenantResponse wraps the tenant list for the backoffice API.
type AllTenantResponse struct {
	Tenants []Tenant `json:"tenants"`
}

// newTenant maps a generated row to the API struct.
func newTenant(id uuid.UUID, slug, name string, status int16, createdAt, updatedAt time.Time) Tenant {
	return Tenant{
		ID:        id,
		Slug:      slug,
		Name:      name,
		Status:    status,
		CreatedAt: createdAt.Unix(),
		UpdatedAt: updatedAt.Unix(),
	}
}
