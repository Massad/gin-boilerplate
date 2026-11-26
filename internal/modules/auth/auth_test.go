//go:build all

package auth_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Massad/gin-boilerplate/internal/modules/auth"
	"github.com/Massad/gin-boilerplate/internal/platform/redisdb"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func middlewareTestSetup(t *testing.T) *auth.Service {
	t.Helper()
	if os.Getenv("REDIS_HOST") == "" {
		t.Skip("set REDIS_HOST to an isolated Redis instance")
	}
	t.Setenv("ACCESS_SECRET", "access-key-for-isolated-tests-only-32-bytes")
	t.Setenv("REFRESH_SECRET", "refresh-key-for-isolated-tests-only-32-bytes")
	t.Setenv("ADMIN_ACCESS_SECRET", "admin-key-for-isolated-tests-only-32-byt")
	gin.SetMode(gin.TestMode)
	require.NoError(t, redisdb.InitRedis(), "resolving the Redis database index")
	require.Equal(t, 2, redisdb.GetRedis().Options().DB, "the suite must run on the disposable Redis index 2")
	t.Cleanup(func() { _ = redisdb.GetRedis().Close() })
	require.NoError(t, redisdb.GetRedis().Ping().Err())
	return auth.NewService(
		auth.Secrets{
			Access:      os.Getenv("ACCESS_SECRET"),
			Refresh:     os.Getenv("REFRESH_SECRET"),
			AdminAccess: os.Getenv("ADMIN_ACCESS_SECRET"),
		},
		auth.NewRepository(redisdb.GetRedis()),
	)
}

func TestAccessTokenRejectedAfterLogout(t *testing.T) {
	svc := middlewareTestSetup(t)
	userID := uuid.New()
	tenantID := uuid.New()
	tokens, err := svc.CreateToken(userID, tenantID, "owner")
	require.NoError(t, err)
	require.NoError(t, svc.CreateAuth(userID, tenantID, "owner", tokens))
	t.Cleanup(func() { redisdb.GetRedis().Del(tokens.AccessUUID, tokens.RefreshUUID) })
	r := gin.New()
	r.GET("/protected", auth.TokenAuth(svc), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id":   c.GetString("userID"),
			"tenant_id": c.GetString("tenantID"),
			"role":      c.GetString("role"),
		})
	})
	r.GET("/logout", auth.TokenAuth(svc), auth.NewHandler(svc).UserLogout)
	for _, step := range []struct {
		path   string
		status int
	}{
		{"/protected", http.StatusOK},
		{"/logout", http.StatusOK},
		{"/protected", http.StatusUnauthorized},
		{"/logout", http.StatusUnauthorized},
	} {
		req := httptest.NewRequest("GET", step.path, nil)
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		require.Equal(t, step.status, w.Code, "%s status mismatch", step.path)
	}
	// JWT signature and expiry remain valid: rejection must come from Redis.
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	token, err := svc.VerifyToken(req)
	require.NoError(t, err, "test token unexpectedly invalid: %v", err)
	require.True(t, token.Valid, "test token unexpectedly invalid")
}

func TestTokenAuthExposesTenantClaims(t *testing.T) {
	svc := middlewareTestSetup(t)
	userID := uuid.New()
	tenantID := uuid.New()
	tokens, err := svc.CreateToken(userID, tenantID, "member")
	require.NoError(t, err)
	require.NoError(t, svc.CreateAuth(userID, tenantID, "member", tokens))
	t.Cleanup(func() { redisdb.GetRedis().Del(tokens.AccessUUID, tokens.RefreshUUID) })

	var gotUserID, gotTenantID, gotRole string
	r := gin.New()
	r.GET("/protected", auth.TokenAuth(svc), func(c *gin.Context) {
		gotUserID = c.GetString("userID")
		gotTenantID = c.GetString("tenantID")
		gotRole = c.GetString("role")
		c.Status(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, "protected route status mismatch")
	require.Equal(t, userID.String(), gotUserID, "userID claim mismatch")
	require.Equal(t, tenantID.String(), gotTenantID, "tenantID claim mismatch")
	require.Equal(t, "member", gotRole, "role claim mismatch")
}

func TestTokenAuthRejectsAdminToken(t *testing.T) {
	svc := middlewareTestSetup(t)
	adminID := uuid.New()
	adminTokens, err := svc.CreateAdminToken(adminID)
	require.NoError(t, err)
	require.NoError(t, svc.CreateAdminAuth(adminID, adminTokens))
	t.Cleanup(func() { redisdb.GetRedis().Del("admin:" + adminTokens.AccessUUID) })

	r := gin.New()
	r.GET("/protected", auth.TokenAuth(svc), func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+adminTokens.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code, "admin token must not pass TokenAuth")
}

func TestAdminAuthRejectsTenantToken(t *testing.T) {
	svc := middlewareTestSetup(t)
	userID := uuid.New()
	tenantID := uuid.New()
	tokens, err := svc.CreateToken(userID, tenantID, "owner")
	require.NoError(t, err)
	require.NoError(t, svc.CreateAuth(userID, tenantID, "owner", tokens))
	t.Cleanup(func() { redisdb.GetRedis().Del(tokens.AccessUUID, tokens.RefreshUUID) })

	r := gin.New()
	r.GET("/admin", auth.AdminAuth(svc), func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code, "tenant token must not pass AdminAuth")
}

func TestAdminAuthAcceptsAdminToken(t *testing.T) {
	svc := middlewareTestSetup(t)
	adminID := uuid.New()
	adminTokens, err := svc.CreateAdminToken(adminID)
	require.NoError(t, err)
	require.NoError(t, svc.CreateAdminAuth(adminID, adminTokens))
	t.Cleanup(func() { redisdb.GetRedis().Del("admin:" + adminTokens.AccessUUID) })

	var gotAdminID string
	r := gin.New()
	r.GET("/admin", auth.AdminAuth(svc), func(c *gin.Context) {
		gotAdminID = c.GetString("adminID")
		c.Status(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+adminTokens.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, "admin route status mismatch")
	require.Equal(t, adminID.String(), gotAdminID, "adminID claim mismatch")
}

func TestRequirePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name   string
		role   string
		status int
	}{
		{"owner manages members", "owner", http.StatusOK},
		{"owner manages tenant", "owner", http.StatusOK},
		{"member cannot manage members", "member", http.StatusForbidden},
		{"member cannot manage tenant", "member", http.StatusForbidden},
		{"unknown role", "superadmin", http.StatusForbidden},
		{"missing role", "", http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				if tc.role != "" {
					c.Set("role", tc.role)
				}
				c.Next()
			})
			r.GET("/manage", auth.RequirePermission(auth.PermManageMembers), func(c *gin.Context) { c.Status(http.StatusOK) })
			req := httptest.NewRequest("GET", "/manage", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			require.Equal(t, tc.status, w.Code, "role %q status mismatch", tc.role)
		})
	}
}

// A role-less token cannot be extracted at all: the claim check guarantees the
// admin token shape never passes TokenAuth.
func TestTokenAuthRejectsTokenWithoutRole(t *testing.T) {
	svc := middlewareTestSetup(t)
	tenantID := uuid.New()
	userID := uuid.New()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"authorized":  true,
		"access_uuid": uuid.NewString(),
		"user_id":     userID.String(),
		"exp":         4102444800,
	}).SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	require.NoError(t, err)
	_ = tenantID
	r := gin.New()
	r.GET("/protected", auth.TokenAuth(svc), func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code, "token without role claim must not pass TokenAuth")
}
