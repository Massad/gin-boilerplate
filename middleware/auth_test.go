//go:build all

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Massad/gin-boilerplate/controllers"
	"github.com/Massad/gin-boilerplate/db"
	"github.com/Massad/gin-boilerplate/middleware"
	"github.com/Massad/gin-boilerplate/models"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func middlewareTestSetup(t *testing.T) *models.AuthModel {
	t.Helper()
	if os.Getenv("REDIS_HOST") == "" {
		t.Skip("set REDIS_HOST to an isolated Redis instance")
	}
	t.Setenv("ACCESS_SECRET", "access-key-for-isolated-tests-only-32-bytes")
	t.Setenv("REFRESH_SECRET", "refresh-key-for-isolated-tests-only-32-bytes")
	t.Setenv("ADMIN_ACCESS_SECRET", "admin-key-for-isolated-tests-only-32-byt")
	gin.SetMode(gin.TestMode)
	db.InitRedis(1)
	t.Cleanup(func() { _ = db.GetRedis().Close() })
	if err := db.GetRedis().Ping().Err(); err != nil {
		t.Fatal(err)
	}
	return new(models.AuthModel)
}

func TestAccessTokenRejectedAfterLogout(t *testing.T) {
	auth := middlewareTestSetup(t)
	userID := uuid.New()
	tenantID := uuid.New()
	tokens, err := auth.CreateToken(userID, tenantID, "owner")
	if err != nil {
		t.Fatal(err)
	}
	if err := auth.CreateAuth(userID, tenantID, "owner", tokens); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.GetRedis().Del(tokens.AccessUUID, tokens.RefreshUUID) })
	r := gin.New()
	r.GET("/protected", middleware.TokenAuth(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id":   c.GetString("userID"),
			"tenant_id": c.GetString("tenantID"),
			"role":      c.GetString("role"),
		})
	})
	r.GET("/logout", middleware.TokenAuth(), new(controllers.UserController).Logout)
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
		if w.Code != step.status {
			t.Fatalf("%s returned %d, want %d", step.path, w.Code, step.status)
		}
	}
	// JWT signature and expiry remain valid: rejection must come from Redis.
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	if token, err := auth.VerifyToken(req); err != nil || !token.Valid {
		t.Fatalf("test token unexpectedly invalid: %v", err)
	}
}

func TestTokenAuthExposesTenantClaims(t *testing.T) {
	auth := middlewareTestSetup(t)
	userID := uuid.New()
	tenantID := uuid.New()
	tokens, err := auth.CreateToken(userID, tenantID, "member")
	if err != nil {
		t.Fatal(err)
	}
	if err := auth.CreateAuth(userID, tenantID, "member", tokens); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.GetRedis().Del(tokens.AccessUUID, tokens.RefreshUUID) })

	var gotUserID, gotTenantID, gotRole string
	r := gin.New()
	r.GET("/protected", middleware.TokenAuth(), func(c *gin.Context) {
		gotUserID = c.GetString("userID")
		gotTenantID = c.GetString("tenantID")
		gotRole = c.GetString("role")
		c.Status(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if gotUserID != userID.String() || gotTenantID != tenantID.String() || gotRole != "member" {
		t.Fatalf("context claims = (%s, %s, %s), want (%s, %s, member)", gotUserID, gotTenantID, gotRole, userID, tenantID)
	}
}

func TestTokenAuthRejectsAdminToken(t *testing.T) {
	auth := middlewareTestSetup(t)
	adminID := uuid.New()
	adminTokens, err := auth.CreateAdminToken(adminID)
	if err != nil {
		t.Fatal(err)
	}
	if err := auth.CreateAdminAuth(adminID, adminTokens); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.GetRedis().Del("admin:" + adminTokens.AccessUUID) })

	r := gin.New()
	r.GET("/protected", middleware.TokenAuth(), func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+adminTokens.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("admin token passed TokenAuth: %d, want 401", w.Code)
	}
}

func TestAdminAuthRejectsTenantToken(t *testing.T) {
	auth := middlewareTestSetup(t)
	userID := uuid.New()
	tenantID := uuid.New()
	tokens, err := auth.CreateToken(userID, tenantID, "owner")
	if err != nil {
		t.Fatal(err)
	}
	if err := auth.CreateAuth(userID, tenantID, "owner", tokens); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.GetRedis().Del(tokens.AccessUUID, tokens.RefreshUUID) })

	r := gin.New()
	r.GET("/admin", middleware.AdminAuth(), func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("tenant token passed AdminAuth: %d, want 401", w.Code)
	}
}

func TestAdminAuthAcceptsAdminToken(t *testing.T) {
	auth := middlewareTestSetup(t)
	adminID := uuid.New()
	adminTokens, err := auth.CreateAdminToken(adminID)
	if err != nil {
		t.Fatal(err)
	}
	if err := auth.CreateAdminAuth(adminID, adminTokens); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.GetRedis().Del("admin:" + adminTokens.AccessUUID) })

	var gotAdminID string
	r := gin.New()
	r.GET("/admin", middleware.AdminAuth(), func(c *gin.Context) {
		gotAdminID = c.GetString("adminID")
		c.Status(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+adminTokens.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if gotAdminID != adminID.String() {
		t.Fatalf("adminID = %s, want %s", gotAdminID, adminID)
	}
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
			r.GET("/manage", middleware.RequirePermission(middleware.PermManageMembers), func(c *gin.Context) { c.Status(http.StatusOK) })
			req := httptest.NewRequest("GET", "/manage", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.status {
				t.Fatalf("role %q: status = %d, want %d", tc.role, w.Code, tc.status)
			}
		})
	}
}

// A role-less token cannot be extracted at all: the claim check guarantees the
// admin token shape never passes TokenAuth.
func TestTokenAuthRejectsTokenWithoutRole(t *testing.T) {
	middlewareTestSetup(t)
	tenantID := uuid.New()
	userID := uuid.New()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"authorized":  true,
		"access_uuid": uuid.NewString(),
		"user_id":     userID.String(),
		"exp":         4102444800,
	}).SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		t.Fatal(err)
	}
	_ = tenantID
	r := gin.New()
	r.GET("/protected", middleware.TokenAuth(), func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("token without role claim passed: %d, want 401", w.Code)
	}
}
