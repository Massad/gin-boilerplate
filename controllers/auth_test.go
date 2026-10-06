//go:build all

package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Massad/gin-boilerplate/db"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func TestRefreshOnlyConsumesMatchingOwner(t *testing.T) {
	if os.Getenv("REDIS_HOST") == "" {
		t.Skip("set REDIS_HOST to an isolated Redis instance")
	}
	t.Setenv("ACCESS_SECRET", "access-key-for-isolated-tests-only-32-bytes")
	t.Setenv("REFRESH_SECRET", "refresh-key-for-isolated-tests-only-32-bytes")
	gin.SetMode(gin.TestMode)
	db.InitRedis(1)
	t.Cleanup(func() { _ = db.GetRedis().Close() })
	if err := db.GetRedis().Ping().Err(); err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.POST("/refresh", new(AuthController).Refresh)
	for _, tc := range []struct {
		name          string
		claimedOwner  int64
		wantStatus    int
		wantRemaining int64
	}{
		{"owner mismatch", 99, http.StatusUnauthorized, 1},
		{"matching owner", 42, http.StatusOK, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			id := uuid.NewString()
			if err := db.GetRedis().Set(id, "42", time.Minute).Err(); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { db.GetRedis().Del(id) })
			token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"refresh_uuid": id, "user_id": tc.claimedOwner, "exp": time.Now().Add(time.Minute).Unix(),
			}).SignedString([]byte(os.Getenv("REFRESH_SECRET")))
			if err != nil {
				t.Fatal(err)
			}
			body, err := json.Marshal(map[string]string{"refresh_token": token})
			if err != nil {
				t.Fatal(err)
			}
			request := func() *httptest.ResponseRecorder {
				req := httptest.NewRequest("POST", "/refresh", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				return w
			}
			w := request()
			if w.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tc.wantStatus)
			}
			remaining, err := db.GetRedis().Exists(id).Result()
			if err != nil {
				t.Fatal(err)
			}
			if remaining != tc.wantRemaining {
				t.Errorf("remaining refresh key = %d, want %d", remaining, tc.wantRemaining)
			}
			if w.Code == http.StatusOK {
				var result map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
					t.Fatal(err)
				}
				for _, kind := range []struct{ name, secret, claim string }{
					{"access_token", "ACCESS_SECRET", "access_uuid"},
					{"refresh_token", "REFRESH_SECRET", "refresh_uuid"},
				} {
					parsed, err := jwt.Parse(result[kind.name], func(_ *jwt.Token) (interface{}, error) {
						return []byte(os.Getenv(kind.secret)), nil
					})
					if err != nil {
						t.Fatal(err)
					}
					claims := parsed.Claims.(jwt.MapClaims)
					db.GetRedis().Del(claims[kind.claim].(string))
					if claims["user_id"] != float64(42) {
						t.Errorf("new %s belongs to %v, want 42", kind.name, claims["user_id"])
					}
				}
				if replay := request(); replay.Code != http.StatusUnauthorized {
					t.Errorf("replay status = %d, want 401", replay.Code)
				}
			}
		})
	}
}
