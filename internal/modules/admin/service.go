// Package admin holds the platform backoffice: the superadmin login (the
// flow of the AdminAuthenticator port certified in the auth module for the
// Phase 7 collage), the tenants listing and the status toggle writing the
// audit line in the same platform transaction (Unit of Work pgdb.WithinTx +
// TenantReader port, design decision 6). Every operation targets the admin
// pool (BYPASSRLS): the tenant table carries no RLS. The CLI bootstrap
// (cmd/api -create-admin) drives the same service.
package admin

import (
	"context"
	"strings"

	auth "github.com/Massad/gin-boilerplate/internal/modules/auth"
	"github.com/Massad/gin-boilerplate/internal/modules/tenant"
	pgdb "github.com/Massad/gin-boilerplate/internal/platform/db"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// TokenService is the technical-service port the login handler consumes
// (consumer-side, design decision 4): issuing and storing the platform
// superadmin session token. The implementation is *auth.Service, injected at
// the composition root.
type TokenService interface {
	CreateAdminToken(adminID uuid.UUID) (*auth.AdminTokenDetails, error)
	CreateAdminAuth(adminID uuid.UUID, td *auth.AdminTokenDetails) error
}

// TenantReader is the tenant access needed by the backoffice (consumer-side
// port): implemented by the tenant module on the platform pool and injected
// at the composition root. The transaction-bound *pgdb.Queries is never a
// parameter or return value of the port.
type TenantReader interface {
	List(ctx context.Context) ([]tenant.Tenant, error)
	UpdateStatusTx(ctx context.Context, tx pgdb.InTx, tenantID uuid.UUID, status int16) error
}

// AdminRepository is the port the service consumes: the platform reads of the
// admin_user table and the in-Tx audit insert. The tx-bound queries are
// reached only inside the module repository.
type AdminRepository interface {
	GetAdminByEmail(ctx context.Context, email string) (pgdb.GetAdminByEmailRow, error)
	CreateAdmin(ctx context.Context, params pgdb.CreateAdminParams) (uuid.UUID, error)
	InsertAdminAuditTx(ctx context.Context, tx pgdb.InTx, params pgdb.InsertAdminAuditParams) error
}

// TxRunner is the Unit-of-Work entry the status-toggle flow drives on the
// platform pool: the pgdb.WithinTx view (design decision 6). pgdb.Store
// satisfies it; the interface exists so the pure suite fakes the transaction
// boundary, the Begin/Rollback/Commit mechanics staying in the platform.
type TxRunner interface {
	WithinTx(ctx context.Context, fn func(tx pgdb.InTx) error) error
}

// Compile-time proofs of the certified port wiring: *auth.Service is the
// TokenService implementation injected at the composition root, the tenant
// repository is the TenantReader implementation.
var (
	_ TokenService = (*auth.Service)(nil)
	_ TenantReader = (*tenant.Repository)(nil)
)

// Service orchestrates the backoffice use cases on its ports: the superadmin
// credential verification, the tenants listing, the status toggle with its
// audit line in the same unit of work, and the CLI upsert.
type Service struct {
	admins  AdminRepository
	tenants TenantReader
	tokens  TokenService
	tx      TxRunner
}

// NewService builds the admin service on the injected repository, tenant
// reader, token service and Unit-of-Work driver.
func NewService(admins AdminRepository, tenants TenantReader, tokens TokenService, tx TxRunner) *Service {
	return &Service{admins: admins, tenants: tenants, tokens: tokens, tx: tx}
}

// normalizeEmail trims the email and lowercases it, mirroring the LOWER($1)
// of the admin_user queries so lookups and upserts agree on one canonical
// email.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// Login verifies the superadmin credentials on the platform pool and returns
// the admin identity on success. The token issuance stays in the handler
// (TokenService port): this flow is the shape of the AdminAuthenticator port
// certified in the auth module for the Phase 7 collage (plan-phases l.65).
func (s *Service) Login(email, password string) (AdminIdentity, error) {
	row, err := s.admins.GetAdminByEmail(context.Background(), normalizeEmail(email))
	if err != nil {
		return AdminIdentity{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(row.Password), []byte(password)); err != nil {
		return AdminIdentity{}, err
	}
	return AdminIdentity{ID: row.ID, Email: row.Email, Name: row.Name}, nil
}

// Create stores a superadmin through the platform pool (the CLI bootstrap
// idempotently resets the password and the name of an existing admin). The
// password arrives already hashed by the caller.
func (s *Service) Create(email string, passwordHash string, name string) (uuid.UUID, error) {
	return s.admins.CreateAdmin(context.Background(), pgdb.CreateAdminParams{
		Email:    normalizeEmail(email),
		Password: passwordHash,
		Name:     name,
	})
}

// AllTenants lists every tenant for the backoffice through the injected
// tenant reader (tenant repository on the platform pool).
func (s *Service) AllTenants() ([]tenant.Tenant, error) {
	return s.tenants.List(context.Background())
}

// SetTenantStatus suspends or reactivates a tenant and writes the audit line
// in the same platform transaction: the flow drives one pgdb.WithinTx unit of
// work with the tenant port and the audit insert on the same InTx.
func (s *Service) SetTenantStatus(adminID, tenantID uuid.UUID, status int16) error {
	ctx := context.Background()
	return s.tx.WithinTx(ctx, func(tx pgdb.InTx) error {
		if err := s.tenants.UpdateStatusTx(ctx, tx, tenantID, status); err != nil {
			return err
		}
		return s.admins.InsertAdminAuditTx(ctx, tx, pgdb.InsertAdminAuditParams{
			AdminID:  adminID,
			TenantID: &tenantID,
			Action:   "tenant.status.set",
		})
	})
}
