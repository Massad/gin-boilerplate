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
)

func TestAccessTokenRejectedAfterLogout(t *testing.T) {
	if os.Getenv("REDIS_HOST") == "" {
		t.Skip("set REDIS_HOST to an isolated Redis instance")
	}
	t.Setenv("ACCESS_SECRET", "access-key-for-isolated-tests-only-32-bytes")
	t.Setenv("REFRESH_SECRET", "refresh-key-for-isolated-tests-only-32-bytes")
	gin.SetMode(gin.TestMode)
	db.InitRedis(1)
	t.Cleanup(func() { _ = db.GetRedis().Close() })
	auth := new(models.AuthModel)
	tokens, err := auth.CreateToken(42)
	if err != nil {
		t.Fatal(err)
	}
	if err := auth.CreateAuth(42, tokens); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.GetRedis().Del(tokens.AccessUUID, tokens.RefreshUUID) })
	r := gin.New()
	r.GET("/protected", middleware.TokenAuth(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user_id": c.GetInt64("userID")})
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
