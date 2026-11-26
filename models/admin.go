package models

import (
	"context"
	"strings"

	"github.com/Massad/gin-boilerplate/internal/modules/tenant"
	pgdb "github.com/Massad/gin-boilerplate/internal/platform/db"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AdminModel ...
type AdminModel struct {
	tenants TenantReader
}

// TenantReader is the tenant access needed by the backoffice (consumer-side
// port of the admin domain): implemented by the tenant module and injected at
// the composition root; it moves into the admin module in Phase 6.
type TenantReader interface {
	List(ctx context.Context) ([]tenant.Tenant, error)
	UpdateStatusTx(ctx context.Context, tx pgdb.InTx, tenantID uuid.UUID, status int16) error
}

// NewAdminModel builds the admin model on the injected tenant reader.
func NewAdminModel(tenants TenantReader) AdminModel {
	return AdminModel{tenants: tenants}
}

// AdminIdentity is the superadmin identity verified at login time. It is
// embedded in the unified login response.
type AdminIdentity struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Name  string    `json:"name"`
}

// Login verifies the superadmin credentials on the platform pool and returns
// the admin identity on success.
func (m AdminModel) Login(email, password string) (AdminIdentity, error) {
	row, err := pgdb.AdminStore().GetAdminByEmail(context.Background(), strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return AdminIdentity{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(row.Password), []byte(password)); err != nil {
		return AdminIdentity{}, err
	}
	return AdminIdentity{ID: row.ID, Email: row.Email, Name: row.Name}, nil
}

// Create stores a superadmin (the CLI bootstrap idempotently resets the
// password and the name of an existing admin).
func (m AdminModel) Create(email string, passwordHash string, name string) (uuid.UUID, error) {
	return pgdb.AdminStore().CreateAdmin(context.Background(), pgdb.CreateAdminParams{
		Email:    strings.ToLower(strings.TrimSpace(email)),
		Password: passwordHash,
		Name:     name,
	})
}

// AllTenants lists every tenant for the backoffice (tenant repository on the
// platform pool).
func (m AdminModel) AllTenants() ([]tenant.Tenant, error) {
	return m.tenants.List(context.Background())
}

// SetTenantStatus suspends or reactivates a tenant and writes the audit line
// in the same platform transaction (the whole model moves into the admin
// module in Phase 6).
func (m AdminModel) SetTenantStatus(adminID, tenantID uuid.UUID, status int16) error {
	ctx := context.Background()
	return pgdb.AdminStore().WithinTx(ctx, func(tx pgdb.InTx) error {
		if err := m.tenants.UpdateStatusTx(ctx, tx, tenantID, status); err != nil {
			return err
		}
		return tx.Queries().InsertAdminAudit(ctx, pgdb.InsertAdminAuditParams{
			AdminID:  adminID,
			TenantID: &tenantID,
			Action:   "tenant.status.set",
		})
	})
}
