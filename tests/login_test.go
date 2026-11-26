//go:build all
// +build all

package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/Massad/gin-boilerplate/internal/modules/admin"
	auth "github.com/Massad/gin-boilerplate/internal/modules/auth"
	user "github.com/Massad/gin-boilerplate/internal/modules/user"
	pgdb "github.com/Massad/gin-boilerplate/internal/platform/db"
	"github.com/Massad/gin-boilerplate/internal/platform/redisdb"
	testutil "github.com/Massad/gin-boilerplate/internal/testutil"
	"github.com/Massad/gin-boilerplate/models"
)

// State shared by the unified-login suite (populated by the tests below, in
// order; the suite runs after the article suite's TestCleanUp).
var unifiedEmail string
var unifiedPassword string
var unifiedTenantID string

// The superadmin fixture of the unified-login suite.
var unifiedAdminEmail string
var unifiedAdminPassword string

/**
* TestUnifiedLoginTenantUser
* A tenant user authenticates through POST /v1/login: kind is tenant_user, the
* identity carries the tenant context and the session half lives in Redis.
*
* Must return response code 200
 */
func TestUnifiedLoginTenantUser(t *testing.T) {
	testRouter := testRouter()

	_, form := testutil.NewUserFixture(t)
	resp := doRequest(testRouter, "POST", "/v1/user/register", user.RegisterForm{
		Name:     form.Name,
		Email:    form.Email,
		Password: form.Password,
		Company:  form.Company,
	}, "")
	require.Equal(t, http.StatusOK, resp.Code)

	unifiedEmail = form.Email
	unifiedPassword = form.Password

	resp = doRequest(testRouter, "POST", "/v1/login", user.LoginForm{Email: unifiedEmail, Password: unifiedPassword}, "")
	require.Equal(t, http.StatusOK, resp.Code)

	var res struct {
		Kind   string             `json:"kind"`
		User   user.User          `json:"user"`
		Tenant user.TenantPayload `json:"tenant"`
		Token  auth.Token         `json:"token"`
	}
	decode(t, resp, &res)

	require.Equal(t, "tenant_user", res.Kind)
	require.Equal(t, unifiedEmail, res.User.Email)
	require.NotEmpty(t, res.User.ID)
	require.NotEmpty(t, res.User.Name)
	require.Equal(t, "owner", res.User.Role)
	require.NotEmpty(t, res.User.TenantID)
	require.Equal(t, res.User.TenantID.String(), res.Tenant.ID)
	require.NotEmpty(t, res.Tenant.Slug)
	require.NotEmpty(t, res.Tenant.Name)
	require.NotEmpty(t, res.Token.AccessToken)
	require.NotEmpty(t, res.Token.RefreshToken)
	unifiedTenantID = res.Tenant.ID

	// The tenant branch advertises both halves of the session.
	token := tokenPayload(t, resp)
	require.Contains(t, token, "access_token")
	require.Contains(t, token, "refresh_token")

	// The access half of the session must live in Redis.
	accessUUID := accessUUIDFromToken(t, res.Token.AccessToken, "ACCESS_SECRET")
	require.Equal(t, int64(1), redisdb.GetRedis().Exists(accessUUID).Val(), "the access token session must exist in Redis")
}

/**
* TestUnifiedLoginSuperadmin
* A superadmin email (no tenant user exists) reaches the admin fallback: kind
* is superadmin and the response carries the stored admin name.
*
* Must return response code 200
 */
func TestUnifiedLoginSuperadmin(t *testing.T) {
	testRouter := testRouter()

	email, password := testutil.NewAdminFixture(t)
	unifiedAdminEmail = email
	unifiedAdminPassword = password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	// The admin fixture stays until TestUnifiedLoginCleanUp: the legacy
	// endpoint tests below reuse it. The superadmin upsert flows through the
	// admin module service (nil ports: the bootstrap never issues tokens,
	// lists tenants or opens a unit of work).
	admins := admin.NewService(admin.NewRepository(pgdb.AdminStore()), nil, nil, nil)
	_, err = admins.Create(email, string(hashed), "Laure Moreau")
	require.NoError(t, err)

	resp := doRequest(testRouter, "POST", "/v1/login", user.LoginForm{Email: email, Password: password}, "")
	require.Equal(t, http.StatusOK, resp.Code)

	var res struct {
		Kind  string `json:"kind"`
		Admin struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"admin"`
		Token struct {
			AccessToken string `json:"access_token"`
		} `json:"token"`
	}
	decode(t, resp, &res)

	require.Equal(t, "superadmin", res.Kind)
	require.NotEmpty(t, res.Admin.ID)
	require.Equal(t, email, res.Admin.Email)
	require.Equal(t, "Laure Moreau", res.Admin.Name)
	require.NotEmpty(t, res.Token.AccessToken)

	// The superadmin branch carries no refresh half: the key is absent from
	// the payload, not serialized empty.
	token := tokenPayload(t, resp)
	require.Contains(t, token, "access_token")
	require.NotContains(t, token, "refresh_token", "the superadmin branch must not advertise a refresh token")

	// The superadmin session half lives under the admin: prefix in Redis.
	adminUUID := accessUUIDFromToken(t, res.Token.AccessToken, "ADMIN_ACCESS_SECRET")
	require.Equal(t, int64(1), redisdb.GetRedis().Exists("admin:"+adminUUID).Val(), "the admin session must exist in Redis")
}

/**
* TestUnifiedLoginWrongPassword
* An existing user with a wrong password fails generically: no admin attempt,
* no session created.
*
* Must return response code 406
 */
func TestUnifiedLoginWrongPassword(t *testing.T) {
	testRouter := testRouter()

	resp := doRequest(testRouter, "POST", "/v1/login", user.LoginForm{Email: unifiedEmail, Password: unifiedPassword + "-wrong"}, "")
	require.Equal(t, http.StatusNotAcceptable, resp.Code)

	var res struct {
		Message string `json:"message"`
	}
	decode(t, resp, &res)
	require.Equal(t, "Invalid login details", res.Message)

	// The same generic failure applies to a superadmin email.
	resp = doRequest(testRouter, "POST", "/v1/login", user.LoginForm{Email: unifiedAdminEmail, Password: unifiedAdminPassword + "-wrong"}, "")
	require.Equal(t, http.StatusNotAcceptable, resp.Code)
}

/**
* TestUnifiedLoginUnknownEmail
* An unknown email fails generically and runs the dummy bcrypt: the response
* time cannot enumerate accounts.
*
* Must return response code 406
 */
func TestUnifiedLoginUnknownEmail(t *testing.T) {
	testRouter := testRouter()

	before := models.DummyHashComparisons()
	resp := doRequest(testRouter, "POST", "/v1/login", user.LoginForm{Email: "nobody-gin-boilerplate@test.com", Password: unifiedPassword}, "")
	require.Equal(t, http.StatusNotAcceptable, resp.Code)
	require.Greater(t, models.DummyHashComparisons(), before, "the unknown-email path must run the dummy bcrypt (timing equalization)")

	var res struct {
		Message string `json:"message"`
	}
	decode(t, resp, &res)
	require.Equal(t, "Invalid login details", res.Message)
}

/**
* TestUnifiedLoginSuspendedTenant
* A suspended tenant stays observable through the unified endpoint: 403.
*
* Must return response code 403
 */
func TestUnifiedLoginSuspendedTenant(t *testing.T) {
	testRouter := testRouter()

	_, err := pgdb.AdminStore().Pool.Exec(context.Background(), "UPDATE tenant SET status=0 WHERE id=$1", unifiedTenantID)
	require.NoError(t, err)

	resp := doRequest(testRouter, "POST", "/v1/login", user.LoginForm{Email: unifiedEmail, Password: unifiedPassword}, "")
	require.Equal(t, http.StatusForbidden, resp.Code)

	_, err = pgdb.AdminStore().Pool.Exec(context.Background(), "UPDATE tenant SET status=1 WHERE id=$1", unifiedTenantID)
	require.NoError(t, err)
}

/**
* TestUnifiedLoginValidation
* The unified endpoint binds user.LoginForm and translates its messages.
*
* Must return response code 406
 */
func TestUnifiedLoginValidation(t *testing.T) {
	testRouter := testRouter()

	resp := doRequest(testRouter, "POST", "/v1/login", user.LoginForm{Email: "invalid@email", Password: unifiedPassword}, "")
	require.Equal(t, http.StatusNotAcceptable, resp.Code)

	var res struct {
		Message string `json:"message"`
	}
	decode(t, resp, &res)
	require.Equal(t, "Please enter a valid email", res.Message)
}

/**
* TestUnifiedLoginLegacyEndpoints
* /v1/user/login and /v1/admin/login stay functional after the unified route.
*
* Must return response code 200
 */
func TestUnifiedLoginLegacyEndpoints(t *testing.T) {
	testRouter := testRouter()

	resp := doRequest(testRouter, "POST", "/v1/user/login", user.LoginForm{Email: unifiedEmail, Password: unifiedPassword}, "")
	require.Equal(t, http.StatusOK, resp.Code)
	// The legacy tenant endpoint keeps its access/refresh pair.
	userToken := tokenPayload(t, resp)
	require.Contains(t, userToken, "access_token")
	require.Contains(t, userToken, "refresh_token")

	resp = doRequest(testRouter, "POST", "/v1/admin/login", user.LoginForm{Email: unifiedAdminEmail, Password: unifiedAdminPassword}, "")
	require.Equal(t, http.StatusOK, resp.Code)
	// The legacy admin endpoint keeps its access-token-only payload.
	adminToken := tokenPayload(t, resp)
	require.Contains(t, adminToken, "access_token")
	require.NotContains(t, adminToken, "refresh_token")
}

/**
* TestUnifiedLoginCleanUp
* Deletes the tenants created by the unified-login suite (cascading to its
* users) and flushes its session keys.
*
* Must pass
 */
func TestUnifiedLoginCleanUp(t *testing.T) {
	_, err := pgdb.AdminStore().Pool.Exec(context.Background(), "DELETE FROM public.tenant WHERE id=$1", unifiedTenantID)
	require.NoError(t, err)

	// The unified-login suite's superadmin fixture (idempotent when absent).
	_, err = pgdb.AdminStore().Pool.Exec(context.Background(), "DELETE FROM public.admin_user WHERE email=$1", unifiedAdminEmail)
	require.NoError(t, err)

	// Flush this suite's tenant/admin session keys (the article suite already
	// cleaned its own keys in TestCleanUp).
	if indexKeys, err := redisdb.GetRedis().Keys("user_sessions:*").Result(); err == nil {
		for _, indexKey := range indexKeys {
			if members, err := redisdb.GetRedis().SMembers(indexKey).Result(); err == nil && len(members) > 0 {
				redisdb.GetRedis().Del(members...)
			}
			redisdb.GetRedis().Del(indexKey)
		}
	}
	if keys, err := redisdb.GetRedis().Keys("admin:*").Result(); err == nil && len(keys) > 0 {
		redisdb.GetRedis().Del(keys...)
	}
}

// tokenPayload returns the raw JSON keys of the "token" object of a login
// response, so a test can assert the presence or the absence of a key.
func tokenPayload(t *testing.T, resp *httptest.ResponseRecorder) map[string]json.RawMessage {
	t.Helper()
	var body struct {
		Token map[string]json.RawMessage `json:"token"`
	}
	decode(t, resp, &body)
	return body.Token
}

// accessUUIDFromToken parses an access token with the given secret env key and
// returns its access_uuid claim.
func accessUUIDFromToken(t *testing.T, token string, secretEnv string) string {
	t.Helper()
	parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv(secretEnv)), nil
	})
	require.NoError(t, err, "the access token must parse with "+secretEnv)
	claims, ok := parsed.Claims.(jwt.MapClaims)
	require.True(t, ok, "the access token claims must be map claims")
	accessUUID, ok := claims["access_uuid"].(string)
	require.True(t, ok, "the access token must carry an access_uuid claim")
	return accessUUID
}
