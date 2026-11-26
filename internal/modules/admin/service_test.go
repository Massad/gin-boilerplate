package admin

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

// fakeAdminRepository stands in for the AdminRepository port.
type fakeAdminRepository struct {
	row    pgdb.GetAdminByEmailRow
	rowErr error
	// loginEmail captures the email passed to the credential lookup.
	loginEmail string

	createID     uuid.UUID
	createErr    error
	createParams pgdb.CreateAdminParams

	auditParams pgdb.InsertAdminAuditParams
	auditErr    error
	auditTx     pgdb.InTx
	auditCalls  int
	// order records its flow step; the tenant-reader fake appends to the
	// same slice, so a test asserts the whole toggle sequence.
	order *[]string
}

func (f *fakeAdminRepository) GetAdminByEmail(_ context.Context, email string) (pgdb.GetAdminByEmailRow, error) {
	f.loginEmail = email
	return f.row, f.rowErr
}

func (f *fakeAdminRepository) CreateAdmin(_ context.Context, params pgdb.CreateAdminParams) (uuid.UUID, error) {
	f.createParams = params
	return f.createID, f.createErr
}

func (f *fakeAdminRepository) InsertAdminAuditTx(_ context.Context, tx pgdb.InTx, params pgdb.InsertAdminAuditParams) error {
	f.auditTx = tx
	f.auditParams = params
	f.auditCalls++
	record(f.order, "audit")
	return f.auditErr
}

// fakeTenantReader stands in for the TenantReader port.
type fakeTenantReader struct {
	listed  []tenant.Tenant
	listErr error

	statusTenantID uuid.UUID
	status         int16
	statusErr      error
	statusTx       pgdb.InTx
	// order records its flow step on the shared toggle sequence.
	order *[]string
}

func (f *fakeTenantReader) List(_ context.Context) ([]tenant.Tenant, error) {
	return f.listed, f.listErr
}

func (f *fakeTenantReader) UpdateStatusTx(_ context.Context, tx pgdb.InTx, tenantID uuid.UUID, status int16) error {
	f.statusTx = tx
	f.statusTenantID = tenantID
	f.status = status
	record(f.order, "status")
	return f.statusErr
}

// record appends a flow step to the shared sequence when the fake carries a
// recorder (the failure-path tests may leave it nil).
func record(order *[]string, step string) {
	if order == nil {
		return
	}
	*order = append(*order, step)
}

// fakeTokens stands in for the TokenService port.
type fakeTokens struct {
	td       *auth.AdminTokenDetails
	tokenErr error
	authErr  error
	// issued reports any token consumption: the login flow keeps the token
	// issuance in the handler (plan-phases l.58), the service Login never
	// touches the port.
	issued bool
}

func (f *fakeTokens) CreateAdminToken(_ uuid.UUID) (*auth.AdminTokenDetails, error) {
	f.issued = true
	return f.td, f.tokenErr
}

func (f *fakeTokens) CreateAdminAuth(_ uuid.UUID, _ *auth.AdminTokenDetails) error {
	f.issued = true
	return f.authErr
}

// fakeTxRunner drives the toggle closure synchronously, like an executed
// transaction commit; the fake InTx is transparent (the module fakes never
// reach the tx-bound queries — data flows through the module ports).
type fakeTxRunner struct {
	ran int
	tx  pgdb.InTx
}

type fakeInTx struct{}

func (fakeInTx) Queries() *pgdb.Queries { return nil }

func (f *fakeTxRunner) WithinTx(_ context.Context, fn func(tx pgdb.InTx) error) error {
	f.ran++
	tx := &fakeInTx{}
	f.tx = tx
	return fn(tx)
}

func adminTokenDetails() *auth.AdminTokenDetails {
	return &auth.AdminTokenDetails{AccessToken: "at", AccessUUID: "au", AtExpires: 42}
}

func mustHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	return string(hash)
}

func TestServiceLoginSuccess(t *testing.T) {
	adminID := uuid.New()
	password := "correct-horse-battery"
	admins := &fakeAdminRepository{
		row: pgdb.GetAdminByEmailRow{ID: adminID, Email: "root@platform.test", Password: mustHash(t, password), Name: "Root"},
	}
	tokens := &fakeTokens{td: adminTokenDetails()}
	svc := NewService(admins, &fakeTenantReader{}, tokens, &fakeTxRunner{})

	admin, err := svc.Login("  Root@Platform.TEST ", password)

	require.NoError(t, err)
	// The email is normalized (trimmed + lowercased) before the lookup.
	require.Equal(t, "root@platform.test", admins.loginEmail)
	require.Equal(t, adminID, admin.ID)
	require.Equal(t, "root@platform.test", admin.Email)
	require.Equal(t, "Root", admin.Name)
	// The token issuance stays in the handler (TokenService port): the
	// service Login is the AdminAuthenticator flow, credentials only.
	require.False(t, tokens.issued)
}

func TestServiceLoginUnknownEmailPropagatesLookupError(t *testing.T) {
	admins := &fakeAdminRepository{rowErr: pgx.ErrNoRows}
	svc := NewService(admins, &fakeTenantReader{}, &fakeTokens{}, &fakeTxRunner{})

	_, err := svc.Login("nobody@platform.test", "whatever")

	// The unified login maps any admin failure generically: the raw lookup
	// error propagates (descendant of AdminModel.Login).
	require.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestServiceLoginWrongPassword(t *testing.T) {
	admins := &fakeAdminRepository{
		row: pgdb.GetAdminByEmailRow{ID: uuid.New(), Email: "root@platform.test", Password: mustHash(t, "stored-password"), Name: "Root"},
	}
	svc := NewService(admins, &fakeTenantReader{}, &fakeTokens{}, &fakeTxRunner{})

	_, err := svc.Login("root@platform.test", "wrong-password")

	require.Error(t, err)
}

func TestServiceAllTenantsPassthrough(t *testing.T) {
	listed := []tenant.Tenant{{ID: uuid.New(), Slug: "acme", Name: "Acme", Status: 1}}
	tenants := &fakeTenantReader{listed: listed}
	svc := NewService(&fakeAdminRepository{}, tenants, &fakeTokens{}, &fakeTxRunner{})

	result, err := svc.AllTenants()

	require.NoError(t, err)
	require.Equal(t, listed, result)
}

func TestServiceAllTenantsErrorPropagates(t *testing.T) {
	listErr := bcrypt.ErrMismatchedHashAndPassword
	tenants := &fakeTenantReader{listErr: listErr}
	svc := NewService(&fakeAdminRepository{}, tenants, &fakeTokens{}, &fakeTxRunner{})

	_, err := svc.AllTenants()

	require.ErrorIs(t, err, listErr)
}

func TestServiceSetTenantStatusOrderAndSingleTx(t *testing.T) {
	order := []string{}
	admins := &fakeAdminRepository{order: &order}
	tenants := &fakeTenantReader{order: &order}
	tx := &fakeTxRunner{}
	svc := NewService(admins, tenants, &fakeTokens{}, tx)

	adminID := uuid.New()
	tenantID := uuid.New()
	require.NoError(t, svc.SetTenantStatus(adminID, tenantID, 0))

	// The audit line is written AFTER the status update...
	require.Equal(t, []string{"status", "audit"}, order)
	// ...inside the ONE transaction the service drove: the same InTx
	// traverses both operations.
	require.Equal(t, 1, tx.ran)
	require.True(t, tenants.statusTx == admins.auditTx, "the audit insert must run on the same InTx as the status update")
	require.Equal(t, tenantID, tenants.statusTenantID)
	require.Equal(t, int16(0), tenants.status)
	require.Equal(t, adminID, admins.auditParams.AdminID)
	require.NotNil(t, admins.auditParams.TenantID)
	require.Equal(t, tenantID, *admins.auditParams.TenantID)
	require.Equal(t, "tenant.status.set", admins.auditParams.Action)
}

func TestServiceSetTenantStatusUpdateFailureSkipsAudit(t *testing.T) {
	order := []string{}
	statusErr := pgx.ErrNoRows
	tenants := &fakeTenantReader{statusErr: statusErr, order: &order}
	admins := &fakeAdminRepository{order: &order}
	svc := NewService(admins, tenants, &fakeTokens{}, &fakeTxRunner{})

	err := svc.SetTenantStatus(uuid.New(), uuid.New(), 1)

	// A failed status update never reaches the audit insert.
	require.ErrorIs(t, err, statusErr)
	require.Equal(t, []string{"status"}, order)
	require.Equal(t, 0, admins.auditCalls)
}

func TestServiceSetTenantStatusAuditFailurePropagates(t *testing.T) {
	auditErr := pgx.ErrNoRows
	admins := &fakeAdminRepository{auditErr: auditErr}
	svc := NewService(admins, &fakeTenantReader{}, &fakeTokens{}, &fakeTxRunner{})

	err := svc.SetTenantStatus(uuid.New(), uuid.New(), 1)

	// The audit failure fails the unit of work: the real WithinTx rolls the
	// status update back with it.
	require.ErrorIs(t, err, auditErr)
	require.Equal(t, 1, admins.auditCalls)
}

func TestServiceCreateUpsertNormalizedEmail(t *testing.T) {
	adminID := uuid.New()
	admins := &fakeAdminRepository{createID: adminID}
	svc := NewService(admins, &fakeTenantReader{}, &fakeTokens{}, &fakeTxRunner{})

	id, err := svc.Create("  Root@Platform.TEST ", "hashed-password", "Root")

	require.NoError(t, err)
	require.Equal(t, adminID, id)
	require.Equal(t, pgdb.CreateAdminParams{
		Email:    "root@platform.test",
		Password: "hashed-password",
		Name:     "Root",
	}, admins.createParams)
}

func TestServiceCreateErrorPropagates(t *testing.T) {
	createErr := bcrypt.ErrMismatchedHashAndPassword
	admins := &fakeAdminRepository{createErr: createErr}
	svc := NewService(admins, &fakeTenantReader{}, &fakeTokens{}, &fakeTxRunner{})

	_, err := svc.Create("root@platform.test", "hashed-password", "Root")

	require.ErrorIs(t, err, createErr)
}
