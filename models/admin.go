package models

import (
	"context"
	"strings"

	pgdb "github.com/Massad/gin-boilerplate/internal/platform/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// AdminModel ...
type AdminModel struct{}

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

// AllTenants lists every tenant for the backoffice.
func (m AdminModel) AllTenants() ([]Tenant, error) {
	return tenantModel.All()
}

// SetTenantStatus suspends or reactivates a tenant and writes the audit line in
// the same platform transaction.
func (m AdminModel) SetTenantStatus(adminID, tenantID uuid.UUID, status int16) error {
	return pgdb.AdminStore().InTx(context.Background(), func(tx pgx.Tx, q *pgdb.Queries) error {
		if err := q.UpdateTenantStatus(context.Background(), pgdb.UpdateTenantStatusParams{ID: tenantID, Status: status}); err != nil {
			return err
		}
		return q.InsertAdminAudit(context.Background(), pgdb.InsertAdminAuditParams{
			AdminID:  adminID,
			TenantID: &tenantID,
			Action:   "tenant.status.set",
		})
	})
}
