package user

import (
	"context"

	pgdb "github.com/Massad/gin-boilerplate/internal/platform/db"
	"github.com/google/uuid"
)

// repository implements the user reads and the in-Tx owner creation on the
// two platform pools: the admin pool (BYPASSRLS) carries the platform
// operations (login read, refresh-time re-read, register unit of work),
// while the RLS-scoped tenant pool backs the tenant-scope read.
type repository struct {
	admin  pgdb.Store
	tenant pgdb.Store
}

// NewRepository builds the user repository on the platform stores injected
// at the composition root: the admin pool for the platform operations, the
// tenant pool for the RLS-scoped read.
func NewRepository(admin, tenant pgdb.Store) UserRepository {
	return repository{admin: admin, tenant: tenant}
}

// ForLogin reads the login row of an email on the admin pool: identity,
// bcrypt hash and the tenant status verified before the password.
func (r repository) ForLogin(ctx context.Context, email string) (pgdb.GetUserForLoginRow, error) {
	return r.admin.GetUserForLogin(ctx, email)
}

// GetUserWithTenant re-reads the user through the admin pool for the
// refresh-time verification: the role is re-freshed and the tenant status
// re-checked on every refresh.
func (r repository) GetUserWithTenant(ctx context.Context, userID uuid.UUID) (pgdb.GetUserWithTenantRow, error) {
	return r.admin.GetUserWithTenant(ctx, userID)
}

// One reads the user inside its tenant scope: the RLS context set by the
// tenant pool makes the row invisible without it (fail-closed policies).
func (r repository) One(ctx context.Context, tenantID, userID uuid.UUID) (pgdb.GetUserByIDRow, error) {
	var row pgdb.GetUserByIDRow
	err := r.tenant.WithTenant(ctx, tenantID, func(q *pgdb.Queries) error {
		fetched, err := q.GetUserByID(ctx, userID)
		if err != nil {
			return err
		}
		row = fetched
		return nil
	})
	return row, err
}

// CreateOwnerTx inserts the owner inside the ongoing platform transaction,
// the same InTx the register flow drives. A conflicting email inserts
// nothing: the one-row contract then reports pgx.ErrNoRows.
func (r repository) CreateOwnerTx(ctx context.Context, tx pgdb.InTx, params pgdb.CreateUserOwnerParams) (uuid.UUID, error) {
	return tx.Queries().CreateUserOwner(ctx, params)
}
