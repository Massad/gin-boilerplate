package models

import (
	"context"
	"errors"
	"strings"

	"github.com/Massad/gin-boilerplate/forms"
	auth "github.com/Massad/gin-boilerplate/internal/modules/auth"
	"github.com/Massad/gin-boilerplate/internal/modules/tenant"
	pgdb "github.com/Massad/gin-boilerplate/internal/platform/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserLoginResponse struct {
	User    User          `json:"user"`
	Tenant  TenantPayload `json:"tenant"`
	Token   auth.Token    `json:"token"`
	Message string        `json:"message"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

// TenantPayload is the tenant subset embedded in user auth responses.
type TenantPayload struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// User ...
type User struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	// Password is the bcrypt hash; never serialized to clients.
	Password string    `json:"-"`
	Name     string    `json:"name"`
	TenantID uuid.UUID `json:"tenant_id"`
	Role     string    `json:"role"`
}

// TenantProvisioner creates tenants inside an ongoing platform transaction.
// Consumer-side port of the user domain (design decision 6): implemented by
// the tenant module and injected at the composition root; it moves into the
// user module in Phase 5. The transaction-bound *pgdb.Queries is never a
// parameter or return value of the port.
type TenantProvisioner interface {
	CreateInTx(ctx context.Context, tx pgdb.InTx, name string) (tenant.Tenant, error)
}

// UserModel issues and stores the session tokens of a login through the
// injected session service (transitory window: the login moves into the user
// module in Phase 5).
type UserModel struct {
	auth    *auth.Service
	tenants TenantProvisioner
}

// NewUserModel builds the user model on the injected session service and the
// tenant provisioner (register creates the tenant inside its platform
// transaction).
func NewUserModel(svc *auth.Service, tenants TenantProvisioner) UserModel {
	return UserModel{auth: svc, tenants: tenants}
}

// ErrTenantSuspended is returned when the credentials are correct but the
// tenant has been suspended: the HTTP answer is 403.
var ErrTenantSuspended = errors.New("tenant suspended")

// normalizeEmail trims the email and lowercases it so that the UNIQUE(email)
// constraint cannot be circumvented by case variants.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// Login authenticates a user on the platform pool (the RLS policies are closed
// to anonymous lookups): the tenant status is verified before the password.
func (m UserModel) Login(form forms.LoginForm) (User, TenantPayload, auth.Token, error) {
	ctx := context.Background()

	row, err := pgdb.AdminStore().GetUserForLogin(ctx, normalizeEmail(form.Email))
	if err != nil {
		return User{}, TenantPayload{}, auth.Token{}, err
	}
	if row.TenantStatus != 1 {
		return User{}, TenantPayload{}, auth.Token{}, ErrTenantSuspended
	}

	user := User{ID: row.ID, Email: row.Email, Password: row.Password, Name: row.Name, TenantID: row.TenantID, Role: row.Role}

	//Compare the password form and database if match
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.Password)); err != nil {
		return User{}, TenantPayload{}, auth.Token{}, err
	}

	tenant := TenantPayload{ID: row.TenantID.String(), Slug: row.Slug, Name: row.TenantName}

	//Generate the JWT auth token
	tokenDetails, err := m.auth.CreateToken(user.ID, user.TenantID, user.Role)
	if err != nil {
		return User{}, TenantPayload{}, auth.Token{}, err
	}

	if err = m.auth.CreateAuth(user.ID, user.TenantID, user.Role, tokenDetails); err != nil {
		return User{}, TenantPayload{}, auth.Token{}, err
	}

	token := auth.Token{AccessToken: tokenDetails.AccessToken, RefreshToken: tokenDetails.RefreshToken}

	return user, tenant, token, nil
}

// Register creates the tenant and its owner atomically on the platform pool.
func (m UserModel) Register(form forms.RegisterForm) (User, tenant.Tenant, error) {
	ctx := context.Background()

	bytePassword := []byte(form.Password)
	hashedPassword, err := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
	if err != nil {
		return User{}, tenant.Tenant{}, errors.New("something went wrong, please try again later")
	}

	email := normalizeEmail(form.Email)

	var user User
	var provisionedTenant tenant.Tenant
	err = pgdb.AdminStore().WithinTx(ctx, func(tx pgdb.InTx) error {
		created, err := m.tenants.CreateInTx(ctx, tx, form.Company)
		if err != nil {
			return err
		}
		provisionedTenant = created
		ownerID, err := tx.Queries().CreateUserOwner(ctx, pgdb.CreateUserOwnerParams{
			Email:    email,
			Password: string(hashedPassword),
			Name:     form.Name,
			TenantID: provisionedTenant.ID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("email already exists")
		}
		if err != nil {
			return errors.New("something went wrong, please try again later")
		}
		user = User{ID: ownerID, Email: email, Name: form.Name, TenantID: provisionedTenant.ID, Role: "owner"}
		return nil
	})
	if err != nil {
		return User{}, tenant.Tenant{}, err
	}

	return user, provisionedTenant, nil
}

// UserWithTenantStatus bundles the refresh-time re-read of a user with the
// freshness data the refresh flow needs.
type UserWithTenantStatus struct {
	User       User
	Tenant     TenantPayload
	TenantSlug string
	Status     int16
}

// OneWithTenant re-reads the user through the platform pool: existence, fresh
// role and active tenant are verified by the refresh flow.
func (m UserModel) OneWithTenant(userID uuid.UUID) (UserWithTenantStatus, error) {
	row, err := pgdb.AdminStore().GetUserWithTenant(context.Background(), userID)
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
func (m UserModel) One(tenantID, userID uuid.UUID) (user User, err error) {
	err = pgdb.TenantStore().WithTenant(context.Background(), tenantID, func(q *pgdb.Queries) error {
		row, err := q.GetUserByID(context.Background(), userID)
		if err != nil {
			return err
		}
		user = User{ID: row.ID, Email: row.Email, Name: row.Name, TenantID: tenantID}
		return nil
	})
	return user, err
}
