package tenant

import (
	"context"

	pgdb "github.com/Massad/gin-boilerplate/internal/platform/db"
	"github.com/google/uuid"
)

// Repository implements the tenant data operations on the platform pool: the
// tenant table carries no RLS and tenant work never runs under a tenant
// context (register and backoffice both target the admin pool).
type Repository struct {
	store pgdb.Store
}

// NewRepository builds the tenant repository on the platform store injected
// at the composition root.
func NewRepository(store pgdb.Store) *Repository {
	return &Repository{store: store}
}

// CreateInTx inserts the tenant inside the ongoing platform transaction,
// retrying on slug collisions with the standard candidates (99 attempts). The
// queries bound to the transaction are reached only here, inside the module.
func (r Repository) CreateInTx(ctx context.Context, tx pgdb.InTx, name string) (Tenant, error) {
	return createTenantInTx(ctx, tx.Queries(), name)
}

// List returns every tenant for the platform backoffice.
func (r Repository) List(ctx context.Context) ([]Tenant, error) {
	rows, err := r.store.ListTenants(ctx)
	if err != nil {
		return nil, err
	}
	tenants := []Tenant{}
	for _, row := range rows {
		tenants = append(tenants, newTenant(row.ID, row.Slug, row.Name, row.Status, row.CreatedAt, row.UpdatedAt))
	}
	return tenants, nil
}

// UpdateStatusTx suspends (or reactivates) a tenant inside the ongoing
// platform transaction, so the caller can share it with its own writes.
func (r Repository) UpdateStatusTx(ctx context.Context, tx pgdb.InTx, tenantID uuid.UUID, status int16) error {
	return tx.Queries().UpdateTenantStatus(ctx, pgdb.UpdateTenantStatusParams{ID: tenantID, Status: status})
}
