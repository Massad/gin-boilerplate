package models

import (
	"errors"
	"sync/atomic"

	"github.com/Massad/gin-boilerplate/forms"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// LoginKind discriminates which directory authenticated a unified login.
type LoginKind string

const (
	// LoginKindTenantUser marks a tenant user session.
	LoginKindTenantUser LoginKind = "tenant_user"
	// LoginKindSuperadmin marks a platform superadmin session.
	LoginKindSuperadmin LoginKind = "superadmin"
)

// ErrInvalidCredentials is the single generic failure of the unified login:
// neither the status nor the body nor the response time may reveal whether an
// email is unknown, belongs to a tenant user, or belongs to a superadmin.
var ErrInvalidCredentials = errors.New("invalid login details")

// dummyPasswordHash is a fixed bcrypt digest (DefaultCost, like the stored
// passwords). It is compared against whenever the email matches no tenant
// user, so an unknown email costs exactly one bcrypt like a wrong password.
const dummyPasswordHash = "$2a$10$94FFhZ.MSC3aFhOoaREEq.qnfTqQBnN5B7E71PAqFYlVzyiGDF/fO"

// dummyHashComparisons counts the dummy bcrypt comparisons executed; the
// integration suite observes it to prove the timing-equalization runs.
var dummyHashComparisons atomic.Int64

// DummyHashComparisons reports how many dummy bcrypt comparisons ran since
// process start (test observability of the enumeration mitigation).
func DummyHashComparisons() int64 { return dummyHashComparisons.Load() }

// UnifiedLoginResult is the discriminated payload of UnifiedLogin: exactly one
// of the tenant-user or superadmin session kinds is meaningful.
type UnifiedLoginResult struct {
	Kind LoginKind
	// User, Tenant and Token are set when Kind is LoginKindTenantUser.
	User   User
	Tenant TenantPayload
	Token  Token
	// Admin and AccessToken are set when Kind is LoginKindSuperadmin.
	Admin       AdminIdentity
	AccessToken string
}

// UnifiedLoginResponse is the documented payload of POST /v1/login. The
// handler answers per branch: the tenant-user branch populates user/tenant and
// an access/refresh token pair, the superadmin branch populates admin and a
// single long-lived access token (8h, no refresh).
type UnifiedLoginResponse struct {
	Kind   LoginKind     `json:"kind"`
	User   User          `json:"user,omitempty"`
	Tenant TenantPayload `json:"tenant,omitempty"`
	Admin  AdminIdentity `json:"admin,omitempty"`
	Token  Token         `json:"token"`
}

// UnifiedLogin authenticates against the tenant directory first and falls back
// to the superadmin directory only when the email matches no user. The client
// observes exactly three outcomes: 200 tenant_user, 200 superadmin, or a
// generic failure:
//
//   - user found + wrong password → ErrInvalidCredentials, no admin attempt
//     (an email must never have its two passwords probed);
//   - user found + suspended tenant → ErrTenantSuspended propagates as 403;
//   - unknown email → one dummy bcrypt (constant cost) then the admin path;
//   - admin not found or wrong password → ErrInvalidCredentials.
//
// The no-overlap policy is explicit: a superadmin email must never match a
// tenant user email, so the two directories are never probed for the same
// account.
func UnifiedLogin(form forms.LoginForm) (UnifiedLoginResult, error) {
	user, tenant, token, userErr := (UserModel{}).Login(form)
	switch {
	case userErr == nil:
		return UnifiedLoginResult{Kind: LoginKindTenantUser, User: user, Tenant: tenant, Token: token}, nil
	case errors.Is(userErr, ErrTenantSuspended):
		return UnifiedLoginResult{}, userErr
	case errors.Is(userErr, pgx.ErrNoRows):
		// Unknown email: equalize the cost below, then try the superadmin
		// directory.
	default:
		// Wrong password or an infrastructure failure: never reach the admin
		// directory.
		return UnifiedLoginResult{}, ErrInvalidCredentials
	}

	// The comparison result is deliberately discarded: the password is wrong
	// by construction. The call exists so that the unknown-email response
	// costs one bcrypt exactly like a wrong-password response.
	dummyHashComparisons.Add(1)
	_ = bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte(form.Password))

	admin, adminErr := (AdminModel{}).Login(form.Email, form.Password)
	if adminErr != nil {
		return UnifiedLoginResult{}, ErrInvalidCredentials
	}

	td, err := (AuthModel{}).CreateAdminToken(admin.ID)
	if err != nil {
		return UnifiedLoginResult{}, ErrInvalidCredentials
	}
	if err := (AuthModel{}).CreateAdminAuth(admin.ID, td); err != nil {
		return UnifiedLoginResult{}, ErrInvalidCredentials
	}

	return UnifiedLoginResult{
		Kind:        LoginKindSuperadmin,
		Admin:       admin,
		AccessToken: td.AccessToken,
	}, nil
}
