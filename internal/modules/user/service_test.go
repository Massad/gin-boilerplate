package user

import (
	"context"
	"testing"

	auth "github.com/Massad/gin-boilerplate/internal/modules/auth"
	"github.com/Massad/gin-boilerplate/internal/modules/tenant"
	pgdb "github.com/Massad/gin-boilerplate/internal/platform/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// fakeUserRepository stands in for the UserRepository port.
type fakeUserRepository struct {
	forLoginRow   pgdb.GetUserForLoginRow
	forLoginErr   error
	withTenantRow pgdb.GetUserWithTenantRow
	withTenantErr error
	oneRow        pgdb.GetUserByIDRow
	oneErr        error
	ownerID       uuid.UUID
	ownerErr      error
	ownerTx       pgdb.InTx
	// loginEmail captures the email passed to the login lookup.
	loginEmail string
	// order records its flow step; the provisioner fake appends to the same
	// slice, so a test asserts the whole register sequence.
	order *[]string
}

func (f *fakeUserRepository) ForLogin(_ context.Context, email string) (pgdb.GetUserForLoginRow, error) {
	f.loginEmail = email
	return f.forLoginRow, f.forLoginErr
}

func (f *fakeUserRepository) GetUserWithTenant(_ context.Context, _ uuid.UUID) (pgdb.GetUserWithTenantRow, error) {
	return f.withTenantRow, f.withTenantErr
}

func (f *fakeUserRepository) One(_ context.Context, _ uuid.UUID, _ uuid.UUID) (pgdb.GetUserByIDRow, error) {
	return f.oneRow, f.oneErr
}

func (f *fakeUserRepository) CreateOwnerTx(_ context.Context, tx pgdb.InTx, _ pgdb.CreateUserOwnerParams) (uuid.UUID, error) {
	f.ownerTx = tx
	record(f.order, "owner")
	return f.ownerID, f.ownerErr
}

// fakeProvisioner stands in for the TenantProvisioner port.
type fakeProvisioner struct {
	created tenant.Tenant
	err     error
	// order records its flow step on the shared register sequence.
	order *[]string
}

func (f *fakeProvisioner) CreateInTx(_ context.Context, _ pgdb.InTx, _ string) (tenant.Tenant, error) {
	record(f.order, "provision")
	return f.created, f.err
}

// record appends a register step to the shared sequence when the fake
// carries a recorder (the failure-path tests may leave it nil).
func record(order *[]string, step string) {
	if order == nil {
		return
	}
	*order = append(*order, step)
}

// fakeTokens stands in for the TokenService port.
type fakeTokens struct {
	td        *auth.TokenDetails
	tokenErr  error
	authErr   error
	tokenUser uuid.UUID
	authUser  uuid.UUID
}

func (f *fakeTokens) CreateToken(userID, _ uuid.UUID, _ string) (*auth.TokenDetails, error) {
	f.tokenUser = userID
	return f.td, f.tokenErr
}

func (f *fakeTokens) CreateAuth(userID, _ uuid.UUID, _ string, _ *auth.TokenDetails) error {
	f.authUser = userID
	return f.authErr
}

// fakeTxRunner drives the register closure synchronously, like an executed
// transaction commit; the fake InTx is transparent (the module fakes never
// reach the tx-bound queries — data flows through the module ports).
type fakeTxRunner struct {
	ran bool
}

type fakeInTx struct{}

func (fakeInTx) Queries() *pgdb.Queries { return nil }

func (f *fakeTxRunner) WithinTx(_ context.Context, fn func(tx pgdb.InTx) error) error {
	f.ran = true
	return fn(fakeInTx{})
}

func testService(users UserRepository, tenants TenantProvisioner, tokens TokenService, tx TxRunner) *Service {
	return NewService(users, tenants, tokens, tx)
}

func validLoginRow(userID, tenantID uuid.UUID, passwordHash string) pgdb.GetUserForLoginRow {
	return pgdb.GetUserForLoginRow{
		ID:           userID,
		Email:        "owner@acme.test",
		Name:         "Owner",
		Password:     passwordHash,
		TenantID:     tenantID,
		Role:         "owner",
		Slug:         "acme",
		TenantName:   "Acme",
		TenantStatus: 1,
	}
}

func mustHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	return string(hash)
}

func tokenDetails() *auth.TokenDetails {
	return &auth.TokenDetails{AccessToken: "at", RefreshToken: "rt", AccessUUID: "au", RefreshUUID: "ru"}
}

func TestServiceLoginSuccess(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	password := "correct-horse-battery"
	users := &fakeUserRepository{forLoginRow: validLoginRow(userID, tenantID, mustHash(t, password))}
	tokens := &fakeTokens{td: tokenDetails()}
	svc := testService(users, &fakeProvisioner{}, tokens, &fakeTxRunner{})

	usr, tenant, token, err := svc.Login(LoginForm{Email: "  Owner@Acme.TEST ", Password: password})

	require.NoError(t, err)
	// The email is normalized (trimmed + lowercased) before the lookup: the
	// UNIQUE(email) constraint cannot be circumvented by case variants.
	require.Equal(t, "owner@acme.test", users.loginEmail)
	require.Equal(t, userID, usr.ID)
	require.Equal(t, "owner", usr.Role)
	require.Equal(t, "acme", tenant.Slug)
	require.Equal(t, tenantID.String(), tenant.ID)
	require.Equal(t, auth.Token{AccessToken: "at", RefreshToken: "rt"}, token)
	// The session tokens are issued and stored for the authenticated user.
	require.Equal(t, userID, tokens.tokenUser)
	require.Equal(t, userID, tokens.authUser)
}

func TestServiceLoginUnknownEmailKeepsErrNoRows(t *testing.T) {
	users := &fakeUserRepository{forLoginErr: pgx.ErrNoRows}
	svc := testService(users, &fakeProvisioner{}, &fakeTokens{}, &fakeTxRunner{})

	_, _, _, err := svc.Login(LoginForm{Email: "nobody@acme.test", Password: "whatever"})

	// The unified login discriminates the unknown-email fallback on
	// pgx.ErrNoRows: the login re-read must keep it intact.
	require.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestServiceLoginSuspendedTenantBeforePassword(t *testing.T) {
	// The stored hash does NOT match the form password: the tenant status
	// check must still short-circuit before the bcrypt comparison.
	storedHash := mustHash(t, "a-completely-different-password")
	users := &fakeUserRepository{forLoginRow: validLoginRow(uuid.New(), uuid.New(), storedHash)}
	users.forLoginRow.TenantStatus = 2
	svc := testService(users, &fakeProvisioner{}, &fakeTokens{}, &fakeTxRunner{})

	_, _, _, err := svc.Login(LoginForm{Email: "owner@acme.test", Password: "correct-horse-battery"})

	require.ErrorIs(t, err, ErrTenantSuspended)
}

func TestServiceLoginWrongPassword(t *testing.T) {
	storedHash := mustHash(t, "stored-password")
	users := &fakeUserRepository{forLoginRow: validLoginRow(uuid.New(), uuid.New(), storedHash)}
	svc := testService(users, &fakeProvisioner{}, &fakeTokens{}, &fakeTxRunner{})

	_, _, _, err := svc.Login(LoginForm{Email: "owner@acme.test", Password: "wrong-password"})

	require.Error(t, err)
	require.NotErrorIs(t, err, ErrTenantSuspended)
}

func TestServiceLoginTokenIssuingFailure(t *testing.T) {
	storedHash := mustHash(t, "correct-horse-battery")
	users := &fakeUserRepository{forLoginRow: validLoginRow(uuid.New(), uuid.New(), storedHash)}
	svc := testService(users, &fakeProvisioner{}, &fakeTokens{tokenErr: pgx.ErrNoRows}, &fakeTxRunner{})

	_, _, _, err := svc.Login(LoginForm{Email: "owner@acme.test", Password: "correct-horse-battery"})

	require.Error(t, err)
}

func TestServiceRegisterSuccess(t *testing.T) {
	ownerID := uuid.New()
	provisionedTenant := tenant.Tenant{ID: uuid.New(), Slug: "acme", Name: "Acme", Status: 1}
	order := []string{}
	users := &fakeUserRepository{ownerID: ownerID, order: &order}
	tenants := &fakeProvisioner{created: provisionedTenant, order: &order}
	svc := testService(users, tenants, &fakeTokens{}, &fakeTxRunner{})

	usr, created, err := svc.Register(RegisterForm{
		Name:     "Owner",
		Email:    "  Owner@Acme.TEST ",
		Password: "correct-horse-battery",
		Company:  "Acme",
	})

	require.NoError(t, err)
	require.Equal(t, ownerID, usr.ID)
	require.Equal(t, "owner@acme.test", usr.Email)
	require.Equal(t, "owner", usr.Role)
	require.Equal(t, provisionedTenant.ID, usr.TenantID)
	require.Equal(t, provisionedTenant, created)
	// The owner is created after the tenant, inside the same unit of work.
	require.Equal(t, []string{"provision", "owner"}, order)
}

func TestServiceRegisterDuplicateEmail(t *testing.T) {
	users := &fakeUserRepository{ownerID: uuid.Nil, ownerErr: pgx.ErrNoRows}
	tenants := &fakeProvisioner{created: tenant.Tenant{ID: uuid.New(), Slug: "acme", Name: "Acme", Status: 1}}
	svc := testService(users, tenants, &fakeTokens{}, &fakeTxRunner{})

	_, _, err := svc.Register(RegisterForm{
		Name:     "Owner",
		Email:    "owner@acme.test",
		Password: "correct-horse-battery",
		Company:  "Acme",
	})

	// The conflicting-email translation depends on the ErrNoRows one-row
	// contract of the insert (ON CONFLICT DO NOTHING).
	require.EqualError(t, err, "email already exists")
}

func TestServiceRegisterOwnerInsertFailure(t *testing.T) {
	users := &fakeUserRepository{ownerErr: bcrypt.ErrMismatchedHashAndPassword}
	tenants := &fakeProvisioner{created: tenant.Tenant{ID: uuid.New(), Slug: "acme", Name: "Acme", Status: 1}}
	svc := testService(users, tenants, &fakeTokens{}, &fakeTxRunner{})

	_, _, err := svc.Register(RegisterForm{
		Name:     "Owner",
		Email:    "owner@acme.test",
		Password: "correct-horse-battery",
		Company:  "Acme",
	})

	require.EqualError(t, err, "something went wrong, please try again later")
}

func TestServiceRegisterInvalidSlugPropagates(t *testing.T) {
	users := &fakeUserRepository{}
	svc := testService(users, &fakeProvisioner{err: tenant.ErrInvalidSlug}, &fakeTokens{}, &fakeTxRunner{})

	_, _, err := svc.Register(RegisterForm{
		Name:     "Owner",
		Email:    "owner@acme.test",
		Password: "correct-horse-battery",
		Company:  "123456",
	})

	// The slug errors propagate to the handler for the current 406 mapping.
	require.ErrorIs(t, err, tenant.ErrInvalidSlug)
}

func TestServiceRegisterSlugUnavailablePropagates(t *testing.T) {
	users := &fakeUserRepository{}
	svc := testService(users, &fakeProvisioner{err: tenant.ErrSlugUnavailable}, &fakeTokens{}, &fakeTxRunner{})

	_, _, err := svc.Register(RegisterForm{
		Name:     "Owner",
		Email:    "owner@acme.test",
		Password: "correct-horse-battery",
		Company:  "Acme",
	})

	require.ErrorIs(t, err, tenant.ErrSlugUnavailable)
}

func TestServiceOneWithTenantMapping(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	users := &fakeUserRepository{withTenantRow: pgdb.GetUserWithTenantRow{
		ID:           userID,
		Email:        "owner@acme.test",
		Name:         "Owner",
		TenantID:     tenantID,
		Role:         "owner",
		Slug:         "acme",
		TenantName:   "Acme",
		TenantStatus: 1,
	}}
	svc := testService(users, &fakeProvisioner{}, &fakeTokens{}, &fakeTxRunner{})

	status, err := svc.OneWithTenant(userID)

	require.NoError(t, err)
	require.Equal(t, userID, status.User.ID)
	require.Equal(t, tenantID, status.User.TenantID)
	require.Equal(t, "acme", status.Tenant.Slug)
	require.Equal(t, "acme", status.TenantSlug)
	require.Equal(t, int16(1), status.Status)
}

func TestServiceOneScopedMapping(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	users := &fakeUserRepository{oneRow: pgdb.GetUserByIDRow{ID: userID, Email: "owner@acme.test", Name: "Owner"}}
	svc := testService(users, &fakeProvisioner{}, &fakeTokens{}, &fakeTxRunner{})

	usr, err := svc.One(tenantID, userID)

	require.NoError(t, err)
	require.Equal(t, userID, usr.ID)
	require.Equal(t, tenantID, usr.TenantID)
}
