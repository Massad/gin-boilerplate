package models

import (
	"context"
	"strings"

	pgdb "github.com/Massad/gin-boilerplate/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// AdminModel ...
type AdminModel struct{}

// Login verifies the superadmin credentials on the platform pool and returns
// the admin id on success.
func (m AdminModel) Login(email, password string) (uuid.UUID, error) {
	row, err := pgdb.AdminStore().GetAdminByEmail(context.Background(), strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return uuid.Nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(row.Password), []byte(password)); err != nil {
		return uuid.Nil, err
	}
	return row.ID, nil
}

// Create stores a superadmin (the CLI bootstrap idempotently resets the
// password of an existing admin).
func (m AdminModel) Create(email string, passwordHash string) (uuid.UUID, error) {
	return pgdb.AdminStore().CreateAdmin(context.Background(), pgdb.CreateAdminParams{
		Email:    strings.ToLower(strings.TrimSpace(email)),
		Password: passwordHash,
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
