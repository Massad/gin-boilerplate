package admin

import (
	"context"

	pgdb "github.com/Massad/gin-boilerplate/internal/platform/db"
	"github.com/google/uuid"
)

// repository implements the admin reads on the platform pool (BYPASSRLS):
// the superadmin credential lookup, the CLI upsert and the in-Tx audit
// insert. The tenant backoffice operations flow through the injected
// TenantReader port (tenant module, same pool).
type repository struct {
	admin pgdb.Store
}

// NewRepository builds the admin repository on the platform store injected
// at the composition root.
func NewRepository(store pgdb.Store) AdminRepository {
	return repository{admin: store}
}

// GetAdminByEmail reads the superadmin row of an email on the platform pool:
// identity, bcrypt hash and display name.
func (r repository) GetAdminByEmail(ctx context.Context, email string) (pgdb.GetAdminByEmailRow, error) {
	return r.admin.GetAdminByEmail(ctx, email)
}

// CreateAdmin upserts the superadmin row on the platform pool: the ON
// CONFLICT (email) clause resets the password and the name of an existing
// admin (idempotent bootstrap).
func (r repository) CreateAdmin(ctx context.Context, params pgdb.CreateAdminParams) (uuid.UUID, error) {
	return r.admin.CreateAdmin(ctx, params)
}

// InsertAdminAuditTx writes the audit line inside the ongoing platform
// transaction, the same InTx the status-toggle flow drives. The queries
// bound to the transaction are reached only here, inside the module.
func (r repository) InsertAdminAuditTx(ctx context.Context, tx pgdb.InTx, params pgdb.InsertAdminAuditParams) error {
	return tx.Queries().InsertAdminAudit(ctx, params)
}
