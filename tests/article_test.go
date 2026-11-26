//go:build all
// +build all

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Massad/gin-boilerplate/forms"
	"github.com/Massad/gin-boilerplate/internal/app"
	"github.com/Massad/gin-boilerplate/internal/modules/article"
	auth "github.com/Massad/gin-boilerplate/internal/modules/auth"
	user "github.com/Massad/gin-boilerplate/internal/modules/user"
	pgdb "github.com/Massad/gin-boilerplate/internal/platform/db"
	"github.com/Massad/gin-boilerplate/internal/platform/redisdb"
	testutil "github.com/Massad/gin-boilerplate/internal/testutil"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// testRouter builds a fresh production-shaped router for the HTTP suites: the
// real app.NewRouter wiring, with the suite's template glob and test mode.
func testRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return app.NewRouter(app.Deps{
		TemplateGlob: "../public/html/*",
		Env:          os.Getenv("ENV"),
		Secrets: auth.Secrets{
			Access:      os.Getenv("ACCESS_SECRET"),
			Refresh:     os.Getenv("REFRESH_SECRET"),
			AdminAccess: os.Getenv("ADMIN_ACCESS_SECRET"),
		},
	})
}

// doRequest runs one request against the router and returns the recorder.
func doRequest(r *gin.Engine, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}
		reader = bytes.NewBuffer(data)
	}
	req, err := http.NewRequest(method, path, reader)
	if err != nil {
		panic(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	return resp
}

// decode unmarshals a JSON response body into v.
func decode(t *testing.T, resp *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), v))
}

var loginCookie string

var testEmail = "test-gin-boilerplate@test.com"
var testPassword = "123456"

var accessToken string
var refreshToken string

var articleID string

// Tenant A (owner of the first registered tenant).
var tenantSlugA string
var tenantIDA string

// Tenant B (second tenant used for cross-tenant isolation checks).
var testEmailB = "tenant-b-gin-boilerplate@test.com"
var tenantSlugB string
var tenantIDB string
var accessTokenB string
var refreshTokenB string

// Platform backoffice admin seeded once by TestIntDB.
var testAdminEmail = "platform-backoffice@test.com"
var testAdminPassword = "admin-strong-password"
var adminAccessToken string

/**
* TestIntDB
* It tests the connection to the database and init the db for this test
*
* Must pass
 */
func TestIntDB(t *testing.T) {

	//Load the .env file
	err := godotenv.Load("../.env")
	require.NoError(t, err, "loading .env: please create one in the root directory")

	ctx := context.Background()
	require.NoError(t, pgdb.InitTenant(ctx, os.Getenv), "initializing the tenant pool: if the error says the database does not exist, run make test once to provision artisan_test; never falls back to the dev database")
	require.NoError(t, pgdb.InitAdmin(ctx, os.Getenv), "initializing the admin pool: if the error says the database does not exist, run make test once to provision artisan_test; never falls back to the dev database")
	var suiteDB string
	require.NoError(t, pgdb.AdminStore().Pool.QueryRow(ctx, "SELECT current_database()").Scan(&suiteDB))
	require.Equal(t, "artisan_test", suiteDB, "the suite must run on the disposable database artisan_test")
	require.NoError(t, redisdb.InitRedis(), "resolving the Redis database index")
	require.Equal(t, 2, redisdb.GetRedis().Options().DB, "the suite must run on the disposable Redis index 2")

	// The suite's starting data is declarative: admin_user.yml carries the
	// backoffice admin, while tenant/user/article start empty (the suite
	// populates them through the HTTP flows).
	testutil.LoadFixtures(t)
}

/**
* TestRegister
* Test user registration: the tenant is created with its owner
*
* Must return response code 200
 */
func TestRegister(t *testing.T) {
	testRouter := testRouter()

	registerForm := user.RegisterForm{
		Name:     "testing",
		Email:    testEmail,
		Password: testPassword,
		Company:  "Test Company",
	}

	resp := doRequest(testRouter, "POST", "/v1/user/register", registerForm, "")
	require.Equal(t, http.StatusOK, resp.Code)

	var res struct {
		Message string `json:"message"`
		User    struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Role  string `json:"role"`
		} `json:"user"`
		Tenant struct {
			ID   string `json:"id"`
			Slug string `json:"slug"`
			Name string `json:"name"`
		} `json:"tenant"`
	}
	decode(t, resp, &res)

	tenantIDA = res.Tenant.ID
	tenantSlugA = res.Tenant.Slug

	require.Equal(t, "test-company", res.Tenant.Slug)
	require.Equal(t, "owner", res.User.Role)
}

/**
* TestRegisterReservedSlug
* A company mapping to a reserved route root is rejected
*
* Must return response code 406
 */
func TestRegisterReservedSlug(t *testing.T) {
	testRouter := testRouter()

	registerForm := user.RegisterForm{
		Name:     "testing",
		Email:    "reserved-slug@test.com",
		Password: testPassword,
		Company:  "Admin",
	}

	resp := doRequest(testRouter, "POST", "/v1/user/register", registerForm, "")
	require.Equal(t, http.StatusNotAcceptable, resp.Code)
}

/**
* TestRegisterDuplicateEmail
* The email is globally unique across tenants
*
* Must return response code 406
 */
func TestRegisterDuplicateEmail(t *testing.T) {
	testRouter := testRouter()

	registerForm := user.RegisterForm{
		Name:     "testing",
		Email:    testEmail,
		Password: testPassword,
		Company:  "Another Company",
	}

	resp := doRequest(testRouter, "POST", "/v1/user/register", registerForm, "")
	require.Equal(t, http.StatusNotAcceptable, resp.Code)
}

/**
* TestRegisterInvalidEmail
* Test user registration with invalid email
*
* Must return response code 406
 */
func TestRegisterInvalidEmail(t *testing.T) {
	testRouter := testRouter()

	registerForm := user.RegisterForm{
		Name:     "testing",
		Email:    "invalid@email",
		Password: testPassword,
		Company:  "Test Company",
	}

	resp := doRequest(testRouter, "POST", "/v1/user/register", registerForm, "")
	require.Equal(t, http.StatusNotAcceptable, resp.Code)
}

/**
* TestLogin
* Test user login: the response carries the tenant and the role
* and the access/refresh tokens are stored
*
* Must return response code 200
 */
func TestLogin(t *testing.T) {
	testRouter := testRouter()

	loginForm := user.LoginForm{
		Email:    testEmail,
		Password: testPassword,
	}

	resp := doRequest(testRouter, "POST", "/v1/user/login", loginForm, "")

	var res struct {
		User struct {
			ID       string `json:"id"`
			Role     string `json:"role"`
			TenantID string `json:"tenant_id"`
		} `json:"user"`
		Tenant struct {
			ID   string `json:"id"`
			Slug string `json:"slug"`
			Name string `json:"name"`
		} `json:"tenant"`
		Token struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"token"`
	}
	decode(t, resp, &res)

	accessToken = res.Token.AccessToken
	refreshToken = res.Token.RefreshToken

	require.Equal(t, http.StatusOK, resp.Code)
	require.Equal(t, tenantSlugA, res.Tenant.Slug)
	require.Equal(t, tenantIDA, res.Tenant.ID)
	require.Equal(t, "owner", res.User.Role)
	require.Equal(t, tenantIDA, res.User.TenantID)
}

/**
* TestLoginUppercaseEmail
* The stored email is normalized: a case variant logs in too
*
* Must return response code 200
 */
func TestLoginUppercaseEmail(t *testing.T) {
	testRouter := testRouter()

	loginForm := user.LoginForm{
		Email:    strings.ToUpper(testEmail),
		Password: testPassword,
	}

	resp := doRequest(testRouter, "POST", "/v1/user/login", loginForm, "")
	require.Equal(t, http.StatusOK, resp.Code)
}

/**
* TestInvalidLogin
* Test invalid login
*
* Must return response code 406
 */
func TestInvalidLogin(t *testing.T) {
	testRouter := testRouter()

	loginForm := user.LoginForm{
		Email:    "wrong@email.com",
		Password: testPassword,
	}

	resp := doRequest(testRouter, "POST", "/v1/user/login", loginForm, "")
	require.Equal(t, http.StatusNotAcceptable, resp.Code)
}

/**
* TestCreateArticle
* Test article creation
*
* Must return response code 200
 */
func TestCreateArticle(t *testing.T) {
	testRouter := testRouter()

	form := article.CreateArticleForm{
		Title:   "Testing article title",
		Content: "Testing article content",
	}

	resp := doRequest(testRouter, "POST", "/v1/article", form, accessToken)

	var res struct {
		Message string `json:"message"`
		ID      string `json:"id"`
	}
	decode(t, resp, &res)

	articleID = res.ID

	require.Equal(t, http.StatusOK, resp.Code)
}

/**
* TestListArticles
* Test listing the tenant articles
*
* Must return response code 200 and contain the created article
 */
func TestListArticles(t *testing.T) {
	testRouter := testRouter()

	resp := doRequest(testRouter, "GET", "/v1/articles", nil, accessToken)

	var res struct {
		Results []struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
			Meta struct {
				Total int `json:"total"`
			} `json:"meta"`
		} `json:"results"`
	}
	decode(t, resp, &res)

	require.Equal(t, http.StatusOK, resp.Code)
	require.Len(t, res.Results, 1)
	require.Equal(t, 1, res.Results[0].Meta.Total)
	var found bool
	for _, item := range res.Results[0].Data {
		if item.ID == articleID {
			found = true
		}
	}
	require.True(t, found, "list must contain the created article")
}

/**
* TestGetArticle
* Test getting one article
*
* Must return response code 200
 */
func TestGetArticle(t *testing.T) {
	testRouter := testRouter()

	resp := doRequest(testRouter, "GET", fmt.Sprintf("/v1/article/%s", articleID), nil, accessToken)
	require.Equal(t, http.StatusOK, resp.Code)

	var res struct {
		Data struct {
			ID   string `json:"id"`
			User struct {
				ID string `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}
	decode(t, resp, &res)
	require.Equal(t, articleID, res.Data.ID)
}

/**
* TestGetInvalidArticle
* Test getting invalid article
*
* Must return response code 404
 */
func TestGetInvalidArticle(t *testing.T) {
	testRouter := testRouter()

	resp := doRequest(testRouter, "GET", "/v1/article/invalid", nil, accessToken)
	require.Equal(t, http.StatusNotFound, resp.Code)
}

/**
* TestGetArticleNotLoggedin
* Test getting the article with logged out user
*
* Must return response code 401
 */
func TestGetArticleNotLoggedin(t *testing.T) {
	testRouter := testRouter()

	resp := doRequest(testRouter, "GET", fmt.Sprintf("/v1/article/%s", articleID), nil, "")
	require.Equal(t, http.StatusUnauthorized, resp.Code)
}

/**
* TestGetArticleUnauthorized
* Test getting the article with unauthorized user (wrong or expired access_token)
*
* Must return response code 401
 */
func TestGetArticleUnauthorized(t *testing.T) {
	testRouter := testRouter()

	resp := doRequest(testRouter, "GET", fmt.Sprintf("/v1/article/%s", articleID), nil, "abc123")
	require.Equal(t, http.StatusUnauthorized, resp.Code)
}

/**
* TestUpdateArticle
* Test updating an article
*
* Must return response code 200
 */
func TestUpdateArticle(t *testing.T) {
	testRouter := testRouter()

	form := article.CreateArticleForm{
		Title:   "Testing new article title",
		Content: "Testing new article content",
	}

	resp := doRequest(testRouter, "PUT", fmt.Sprintf("/v1/article/%s", articleID), form, accessToken)
	require.Equal(t, http.StatusOK, resp.Code)
}

/**
* TestCrossTenantIsolation
* Tenant B can neither read, update nor delete the tenant A article; its own
* list stays empty; and it creates and reads its own article.
 */
func TestCrossTenantIsolation(t *testing.T) {
	testRouter := testRouter()

	// Register tenant B.
	registerForm := user.RegisterForm{
		Name:     "testingb",
		Email:    testEmailB,
		Password: testPassword,
		Company:  "Second Company",
	}
	resp := doRequest(testRouter, "POST", "/v1/user/register", registerForm, "")
	require.Equal(t, http.StatusOK, resp.Code)
	var registration struct {
		Tenant struct {
			ID   string `json:"id"`
			Slug string `json:"slug"`
		} `json:"tenant"`
	}
	decode(t, resp, &registration)
	tenantIDB = registration.Tenant.ID
	tenantSlugB = registration.Tenant.Slug
	require.NotEmpty(t, tenantIDB)

	// Login tenant B.
	loginForm := user.LoginForm{Email: testEmailB, Password: testPassword}
	resp = doRequest(testRouter, "POST", "/v1/user/login", loginForm, "")
	var loginRes struct {
		User struct {
			TenantID string `json:"tenant_id"`
		} `json:"user"`
		Tenant struct {
			Slug string `json:"slug"`
		} `json:"tenant"`
		Token struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"token"`
	}
	decode(t, resp, &loginRes)
	accessTokenB = loginRes.Token.AccessToken
	refreshTokenB = loginRes.Token.RefreshToken
	require.Equal(t, http.StatusOK, resp.Code)
	require.Equal(t, tenantIDB, loginRes.User.TenantID)
	require.Equal(t, "second-company", loginRes.Tenant.Slug)

	// A's article is invisible for B (RLS: the row does not exist).
	require.Equal(t, http.StatusNotFound, doRequest(testRouter, "GET", fmt.Sprintf("/v1/article/%s", articleID), nil, accessTokenB).Code)
	require.Equal(t, http.StatusNotFound, doRequest(testRouter, "PUT", fmt.Sprintf("/v1/article/%s", articleID), article.CreateArticleForm{Title: "Stolen title", Content: "Stolen content"}, accessTokenB).Code)
	require.Equal(t, http.StatusNotFound, doRequest(testRouter, "DELETE", fmt.Sprintf("/v1/article/%s", articleID), nil, accessTokenB).Code)

	// B's own list is empty.
	var listRes struct {
		Results []struct {
			Data []string `json:"data"`
			Meta struct {
				Total int `json:"total"`
			} `json:"meta"`
		} `json:"results"`
	}
	decode(t, doRequest(testRouter, "GET", "/v1/articles", nil, accessTokenB), &listRes)
	require.Equal(t, 0, listRes.Results[0].Meta.Total)
	require.Len(t, listRes.Results[0].Data, 0)
}

/**
* TestGetArticleB
* Tenant B reads its own article normally
*
* Must return response code 200
 */
func TestGetArticleB(t *testing.T) {
	testRouter := testRouter()

	form := article.CreateArticleForm{
		Title:   "B article title",
		Content: "B article content",
	}
	resp := doRequest(testRouter, "POST", "/v1/article", form, accessTokenB)
	require.Equal(t, http.StatusOK, resp.Code)

	var created struct {
		ID string `json:"id"`
	}
	decode(t, resp, &created)

	resp = doRequest(testRouter, "GET", fmt.Sprintf("/v1/article/%s", created.ID), nil, accessTokenB)
	require.Equal(t, http.StatusOK, resp.Code)
}

/**
* TestRLSFailClosed
* Raw SQL against the tenant pool proves the row level security is enforced:
* without a tenant context every tenant row is invisible, a wrong tenant sees
* nothing, and a cross tenant insert is rejected.
 */
func TestRLSFailClosed(t *testing.T) {
	ctx := context.Background()
	pool := pgdb.TenantStore().Pool

	// 1. No tenant context: the policies close everything.
	tx, err := pool.Begin(ctx)
	require.NoError(t, err)
	var count int
	err = tx.QueryRow(ctx, `SELECT count(*) FROM article`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count, "article rows must be invisible without tenant context")
	err = tx.QueryRow(ctx, `SELECT count(*) FROM "user"`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count, "user rows must be invisible without tenant context")
	_ = tx.Rollback(ctx)

	// 2. A wrong tenant context: still nothing.
	tx, err = pool.Begin(ctx)
	require.NoError(t, err)
	_, err = tx.Exec(ctx, `SELECT set_config('app.tenant_id', $1, true)`, uuid.NewString())
	require.NoError(t, err)
	err = tx.QueryRow(ctx, `SELECT count(*) FROM article`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count, "article rows must be invisible for a foreign tenant")
	_ = tx.Rollback(ctx)

	// 3. Cross tenant insert is rejected.
	tx, err = pool.Begin(ctx)
	require.NoError(t, err)
	_, err = tx.Exec(ctx, `SELECT set_config('app.tenant_id', $1, true)`, tenantIDB)
	require.NoError(t, err)
	_, err = tx.Exec(ctx,
		`INSERT INTO article (tenant_id, user_id, title, content)
		 SELECT $1, b.id, 'stolen', 'stolen content'
		 FROM "user" b
		 WHERE b.email = $2`,
		tenantIDA, testEmailB)
	// The insert targets tenant A while the RLS context is tenant B: the
	// policy must reject the violating row with 42501.
	require.Error(t, err, "cross-tenant insert unexpectedly succeeded (RLS WITH CHECK failed)")
	pgErr, ok := err.(*pgconn.PgError)
	require.True(t, ok && pgErr.Code == "42501", "expected RLS violation 42501, got %v", err)
	_ = tx.Rollback(ctx)

	// 4. Within its own tenant context, tenant B sees its rows and can insert.
	tx, err = pool.Begin(ctx)
	require.NoError(t, err)
	_, err = tx.Exec(ctx, `SELECT set_config('app.tenant_id', $1, true)`, tenantIDB)
	require.NoError(t, err)
	var bCount int
	err = tx.QueryRow(ctx, `SELECT count(*) FROM "user"`).Scan(&bCount)
	require.NoError(t, err)
	require.Equal(t, 1, bCount, "tenant B sees its own user inside its context")
	err = tx.QueryRow(ctx, `SELECT count(*) FROM article`).Scan(&bCount)
	require.NoError(t, err)
	require.Equal(t, 1, bCount, "tenant B sees exactly its own article")
	_ = tx.Rollback(ctx)
}

/**
* TestOneOwnerConstraint
* The partial unique index guarantees a single owner per tenant.
 */
func TestOneOwnerConstraint(t *testing.T) {
	ctx := context.Background()
	pool := pgdb.TenantStore().Pool

	tx, err := pool.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `SELECT set_config('app.tenant_id', $1, true)`, tenantIDA)
	require.NoError(t, err)
	_, err = tx.Exec(ctx,
		`INSERT INTO "user" (email, password, name, tenant_id, role)
		 VALUES ($1, $2, $3, $4, 'owner')`,
		"second-owner@test.com", "irrelevant-hash", "Second Owner", tenantIDA)
	require.Error(t, err, "second owner unexpectedly inserted")
	pgErr, ok := err.(*pgconn.PgError)
	require.True(t, ok && pgErr.Code == "23505", "expected unique violation 23505, got %v", err)
}

/**
* TestRefreshToken
* Test refreshing the token with valid refresh_token
*
* Must return response code 200
 */
func TestRefreshToken(t *testing.T) {
	testRouter := testRouter()

	tokenForm := forms.Token{RefreshToken: refreshToken}

	resp := doRequest(testRouter, "POST", "/v1/token/refresh", tokenForm, "")

	var res struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	decode(t, resp, &res)

	// The original refresh token is kept as the replay subject: the refreshed
	// pair is a different session.
	require.NotEmpty(t, res.AccessToken)
	require.NotEmpty(t, res.RefreshToken)

	require.Equal(t, http.StatusOK, resp.Code)
}

/**
* TestInvalidRefreshToken
* Test refreshing the token with invalid refresh_token
*
* Must return response code 401
 */
func TestInvalidRefreshToken(t *testing.T) {
	testRouter := testRouter()

	//Since we didn't update it in the test before - this will not be valid anymore
	tokenForm := forms.Token{RefreshToken: refreshToken}

	resp := doRequest(testRouter, "POST", "/v1/token/refresh", tokenForm, "")
	require.Equal(t, http.StatusUnauthorized, resp.Code)
}

/**
* TestUserLogout
* Test logout a user
*
* Must return response code 200
 */
func TestUserLogout(t *testing.T) {
	testRouter := testRouter()

	resp := doRequest(testRouter, "GET", "/v1/user/logout", nil, accessToken)
	require.Equal(t, http.StatusOK, resp.Code)
}

/**
* TestAdminLogin
* Login to the platform backoffice with the seeded admin_user
*
* Must return response code 200
 */
func TestAdminLogin(t *testing.T) {
	testRouter := testRouter()

	loginForm := user.LoginForm{Email: testAdminEmail, Password: testAdminPassword}

	resp := doRequest(testRouter, "POST", "/v1/admin/login", loginForm, "")

	var res struct {
		Token struct {
			AccessToken string `json:"access_token"`
		} `json:"token"`
	}
	decode(t, resp, &res)
	adminAccessToken = res.Token.AccessToken

	require.Equal(t, http.StatusOK, resp.Code)
	require.NotEmpty(t, adminAccessToken)
}

/**
* TestAdminLoginWithUserCredentials
* A tenant user cannot log into the backoffice
*
* Must return response code 406
 */
func TestAdminLoginWithUserCredentials(t *testing.T) {
	testRouter := testRouter()

	loginForm := user.LoginForm{Email: testEmail, Password: testPassword}

	resp := doRequest(testRouter, "POST", "/v1/admin/login", loginForm, "")
	require.Equal(t, http.StatusNotAcceptable, resp.Code)
}

/**
* TestAdminTenants
* The backoffice lists both tenants
*
* Must return response code 200
 */
func TestAdminTenants(t *testing.T) {
	testRouter := testRouter()

	resp := doRequest(testRouter, "GET", "/v1/admin/tenants", nil, adminAccessToken)

	var res struct {
		Tenants []struct {
			ID   string `json:"id"`
			Slug string `json:"slug"`
		} `json:"tenants"`
	}
	decode(t, resp, &res)

	require.Equal(t, http.StatusOK, resp.Code)
	slugs := map[string]bool{}
	for _, tenant := range res.Tenants {
		slugs[tenant.Slug] = true
	}
	require.True(t, slugs[tenantSlugA])
	require.True(t, slugs[tenantSlugB])
}

/**
* TestAdminTokenOnTenantRoute
* A superadmin token must not reach tenant routes
*
* Must return response code 401
 */
func TestAdminTokenOnTenantRoute(t *testing.T) {
	testRouter := testRouter()

	resp := doRequest(testRouter, "GET", "/v1/articles", nil, adminAccessToken)
	require.Equal(t, http.StatusUnauthorized, resp.Code)
}

/**
* TestUserTokenOnAdminRoute
* A tenant token must not reach the backoffice
*
* Must return response code 401
 */
func TestUserTokenOnAdminRoute(t *testing.T) {
	testRouter := testRouter()

	resp := doRequest(testRouter, "GET", "/v1/admin/tenants", nil, accessTokenB)
	require.Equal(t, http.StatusUnauthorized, resp.Code)
}

/**
* TestAdminSuspendTenant
* The backoffice suspends tenant B
*
* Must return response code 200
 */
func TestAdminSuspendTenant(t *testing.T) {
	testRouter := testRouter()

	body := map[string]int16{"status": 0}
	resp := doRequest(testRouter, "POST", fmt.Sprintf("/v1/admin/tenants/%s/status", tenantIDB), body, adminAccessToken)
	require.Equal(t, http.StatusOK, resp.Code)
}

/**
* TestSuspendedTenantLogin
* A suspended tenant rejects logins
*
* Must return response code 403
 */
func TestSuspendedTenantLogin(t *testing.T) {
	testRouter := testRouter()

	loginForm := user.LoginForm{Email: testEmailB, Password: testPassword}
	resp := doRequest(testRouter, "POST", "/v1/user/login", loginForm, "")
	require.Equal(t, http.StatusForbidden, resp.Code)
}

/**
* TestSuspendedTenantRefresh
* A suspended tenant rejects refreshes
*
* Must return response code 401
 */
func TestSuspendedTenantRefresh(t *testing.T) {
	testRouter := testRouter()

	tokenForm := forms.Token{RefreshToken: refreshTokenB}
	resp := doRequest(testRouter, "POST", "/v1/token/refresh", tokenForm, "")
	require.Equal(t, http.StatusUnauthorized, resp.Code)
}

/**
* TestAdminReactivateTenant
* The backoffice reactivates tenant B: logins work again
*
* Must return response code 200
 */
func TestAdminReactivateTenant(t *testing.T) {
	testRouter := testRouter()

	body := map[string]int16{"status": 1}
	resp := doRequest(testRouter, "POST", fmt.Sprintf("/v1/admin/tenants/%s/status", tenantIDB), body, adminAccessToken)
	require.Equal(t, http.StatusOK, resp.Code)

	loginForm := user.LoginForm{Email: testEmailB, Password: testPassword}
	resp = doRequest(testRouter, "POST", "/v1/user/login", loginForm, "")
	require.Equal(t, http.StatusOK, resp.Code)
}

/**
* TestAdminLogout
* The admin session is revoked immediately (no refresh)
*
* Must return response code 200, then the token is dead.
 */
func TestAdminLogout(t *testing.T) {
	testRouter := testRouter()

	resp := doRequest(testRouter, "GET", "/v1/admin/logout", nil, adminAccessToken)
	require.Equal(t, http.StatusOK, resp.Code)

	resp = doRequest(testRouter, "GET", "/v1/admin/tenants", nil, adminAccessToken)
	require.Equal(t, http.StatusUnauthorized, resp.Code)
}

/**
* TestCleanUp
* Deletes the tenants (cascading to users and articles) and the test admin,
* then flushes the test session keys.
*
* Must pass
 */
func TestCleanUp(t *testing.T) {
	ctx := context.Background()

	var suiteDB string
	require.NoError(t, pgdb.AdminStore().Pool.QueryRow(ctx, "SELECT current_database()").Scan(&suiteDB))
	require.Equal(t, "artisan_test", suiteDB, "the suite must run on the disposable database artisan_test")

	_, err := pgdb.AdminStore().Pool.Exec(ctx, "DELETE FROM public.tenant")
	require.NoError(t, err)
	_, err = pgdb.AdminStore().Pool.Exec(ctx, "DELETE FROM public.admin_user WHERE email=$1", testAdminEmail)
	require.NoError(t, err)

	// Flush this suite's Redis keys: tenant session index sets and admin keys.
	if indexKeys, err := redisdb.GetRedis().Keys("user_sessions:*").Result(); err == nil {
		for _, indexKey := range indexKeys {
			if members, err := redisdb.GetRedis().SMembers(indexKey).Result(); err == nil {
				if len(members) > 0 {
					redisdb.GetRedis().Del(members...)
				}
			}
			redisdb.GetRedis().Del(indexKey)
		}
	}
	if keys, err := redisdb.GetRedis().Keys("admin:*").Result(); err == nil && len(keys) > 0 {
		redisdb.GetRedis().Del(keys...)
	}
	redisdb.GetRedis().Del(accessTokenB, refreshTokenB)

	// Let the environment settle before the process exits.
	_ = time.Now()
}
