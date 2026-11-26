// Package user holds the user domain: the user login (tenant status verified
// before the password), the register flow creating the tenant and its owner
// in one platform transaction (Unit of Work pgdb.WithinTx + TenantProvisioner
// port) and the refresh-time re-reads. The platform operations target the
// admin pool (BYPASSRLS); the RLS-scoped reads run on the tenant pool. The
// service carries the flows of the ports UserAuthenticator (Login(form)) and
// UserReader (OneWithTenant) certified in the auth module for the Phase 7
// collage (plan-phases l.39/l.52/l.64).
package user

import (
	"context"
	"errors"
	"strings"

	auth "github.com/Massad/gin-boilerplate/internal/modules/auth"
	"github.com/Massad/gin-boilerplate/internal/modules/tenant"
	pgdb "github.com/Massad/gin-boilerplate/internal/platform/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// TokenService is the technical-service port the login consumes
// (consumer-side, design decision 4): issuing and storing the session tokens
// of a tenant login. The implementation is *auth.Service, injected at the
// composition root.
type TokenService interface {
	CreateToken(userID, tenantID uuid.UUID, role string) (*auth.TokenDetails, error)
	CreateAuth(userID, tenantID uuid.UUID, role string, td *auth.TokenDetails) error
}

// TenantProvisioner creates tenants inside an ongoing platform transaction
// (consumer-side port, design decision 6): implemented by the tenant module
// and injected at the composition root. The transaction-bound *pgdb.Queries
// is never a parameter or return value of the port.
type TenantProvisioner interface {
	CreateInTx(ctx context.Context, tx pgdb.InTx, name string) (tenant.Tenant, error)
}

// UserRepository is the port the service consumes: the platform reads of the
// user rows and the in-Tx owner insert. The tx-bound queries are reached
// only inside the module repository.
type UserRepository interface {
	ForLogin(ctx context.Context, email string) (pgdb.GetUserForLoginRow, error)
	GetUserWithTenant(ctx context.Context, userID uuid.UUID) (pgdb.GetUserWithTenantRow, error)
	One(ctx context.Context, tenantID, userID uuid.UUID) (pgdb.GetUserByIDRow, error)
	CreateOwnerTx(ctx context.Context, tx pgdb.InTx, params pgdb.CreateUserOwnerParams) (uuid.UUID, error)
}

// TxRunner is the Unit-of-Work entry the register flow drives on the
// platform pool: the pgdb.WithinTx view (design decision 6). pgdb.Store
// satisfies it; the interface exists so the pure suite fakes the transaction
// boundary, the Begin/Rollback/Commit mechanics staying in the platform.
type TxRunner interface {
	WithinTx(ctx context.Context, fn func(tx pgdb.InTx) error) error
}

// Compile-time proofs of the certified port wiring: *auth.Service is the
// TokenService implementation injected at the composition root, the tenant
// repository is the TenantProvisioner implementation.
var (
	_ TokenService      = (*auth.Service)(nil)
	_ TenantProvisioner = (*tenant.Repository)(nil)
)

// ErrTenantSuspended is returned when the credentials are correct but the
// tenant has been suspended: the HTTP answer is 403.
var ErrTenantSuspended = errors.New("tenant suspended")

// Service orchestrates the user use cases on its ports: Login (status before
// password), Register (tenant + owner in one unit of work) and the re-reads.
type Service struct {
	users   UserRepository
	tenants TenantProvisioner
	tokens  TokenService
	tx      TxRunner
}

// NewService builds the user service on the injected repository, tenant
// provisioner, token service and Unit-of-Work driver.
func NewService(users UserRepository, tenants TenantProvisioner, tokens TokenService, tx TxRunner) *Service {
	return &Service{users: users, tenants: tenants, tokens: tokens, tx: tx}
}

// normalizeEmail trims the email and lowercases it so that the UNIQUE(email)
// constraint cannot be circumvented by case variants.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// Login authenticates a user on the platform pool (the RLS policies are closed
// to anonymous lookups): the tenant status is verified before the password.
func (s *Service) Login(form LoginForm) (User, TenantPayload, auth.Token, error) {
	ctx := context.Background()

	row, err := s.users.ForLogin(ctx, normalizeEmail(form.Email))
	if err != nil {
		return User{}, TenantPayload{}, auth.Token{}, err
	}
	if row.TenantStatus != 1 {
		return User{}, TenantPayload{}, auth.Token{}, ErrTenantSuspended
	}

	usr := User{ID: row.ID, Email: row.Email, Password: row.Password, Name: row.Name, TenantID: row.TenantID, Role: row.Role}

	//Compare the password form and database if match
	if err := bcrypt.CompareHashAndPassword([]byte(usr.Password), []byte(form.Password)); err != nil {
		return User{}, TenantPayload{}, auth.Token{}, err
	}

	tenantData := TenantPayload{ID: row.TenantID.String(), Slug: row.Slug, Name: row.TenantName}

	//Generate the JWT auth token
	tokenDetails, err := s.tokens.CreateToken(usr.ID, usr.TenantID, usr.Role)
	if err != nil {
		return User{}, TenantPayload{}, auth.Token{}, err
	}

	if err = s.tokens.CreateAuth(usr.ID, usr.TenantID, usr.Role, tokenDetails); err != nil {
		return User{}, TenantPayload{}, auth.Token{}, err
	}

	token := auth.Token{AccessToken: tokenDetails.AccessToken, RefreshToken: tokenDetails.RefreshToken}

	return usr, tenantData, token, nil
}

// Register creates the tenant and its owner atomically on the platform pool:
// the flow drives one pgdb.WithinTx unit of work with the provisioning port
// and the owner insert on the same InTx.
func (s *Service) Register(form RegisterForm) (User, tenant.Tenant, error) {
	ctx := context.Background()

	bytePassword := []byte(form.Password)
	hashedPassword, err := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
	if err != nil {
		return User{}, tenant.Tenant{}, errors.New("something went wrong, please try again later")
	}

	email := normalizeEmail(form.Email)

	var usr User
	var provisioned tenant.Tenant
	err = s.tx.WithinTx(ctx, func(tx pgdb.InTx) error {
		created, err := s.tenants.CreateInTx(ctx, tx, form.Company)
		if err != nil {
			return err
		}
		provisioned = created
		ownerID, err := s.users.CreateOwnerTx(ctx, tx, pgdb.CreateUserOwnerParams{
			Email:    email,
			Password: string(hashedPassword),
			Name:     form.Name,
			TenantID: provisioned.ID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("email already exists")
		}
		if err != nil {
			return errors.New("something went wrong, please try again later")
		}
		usr = User{ID: ownerID, Email: email, Name: form.Name, TenantID: provisioned.ID, Role: "owner"}
		return nil
	})
	if err != nil {
		return User{}, tenant.Tenant{}, err
	}

	return usr, provisioned, nil
}

// OneWithTenant re-reads the user through the platform pool: existence, fresh
// role and active tenant are verified by the refresh flow.
func (s *Service) OneWithTenant(userID uuid.UUID) (UserWithTenantStatus, error) {
	row, err := s.users.GetUserWithTenant(context.Background(), userID)
	if err != nil {
		return UserWithTenantStatus{}, err
	}
	return UserWithTenantStatus{
		User: User{ID: row.ID, Email: row.Email, Name: row.Name, TenantID: row.TenantID, Role: row.Role},
		Tenant: TenantPayload{
			ID:   row.TenantID.String(),
			Slug: row.Slug,
			Name: row.TenantName,
		},
		TenantSlug: row.Slug,
		Status:     row.TenantStatus,
	}, nil
}

// One reads the user inside its tenant scope (the RLS context is required).
func (s *Service) One(tenantID, userID uuid.UUID) (User, error) {
	row, err := s.users.One(context.Background(), tenantID, userID)
	if err != nil {
		return User{}, err
	}
	return User{ID: row.ID, Email: row.Email, Name: row.Name, TenantID: tenantID}, nil
}
