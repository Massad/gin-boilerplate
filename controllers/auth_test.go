//go:build all

package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Massad/gin-boilerplate/db"
	pgdb "github.com/Massad/gin-boilerplate/internal/db"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// TestRefreshOnlyConsumesMatchingOwner exercises the adapted Lua script and the
// platform pool re-verification of the refresh flow: ownership check on the
// Redis value, fresh role from the database, suspended tenant rejection.
func TestRefreshOnlyConsumesMatchingOwner(t *testing.T) {
	if os.Getenv("REDIS_HOST") == "" || os.Getenv("DB_HOST") == "" {
		t.Skip("set REDIS_HOST and DB_HOST to isolated services")
	}
	t.Setenv("ACCESS_SECRET", "access-key-for-isolated-tests-only-32-bytes")
	t.Setenv("REFRESH_SECRET", "refresh-key-for-isolated-tests-only-32-bytes")
	gin.SetMode(gin.TestMode)
	db.InitRedis(1)
	t.Cleanup(func() { _ = db.GetRedis().Close() })
	if err := db.GetRedis().Ping().Err(); err != nil {
		t.Fatal(err)
	}
	if err := pgdb.InitTenant(context.Background(), os.Getenv); err != nil {
		t.Fatal(err)
	}
	if err := pgdb.InitAdmin(context.Background(), os.Getenv); err != nil {
		t.Fatal(err)
	}

	// Seed a tenant and its owner on the platform pool.
	ctx := context.Background()
	email := "refresh-owner@test.com"
	tenant, err := pgdb.AdminStore().CreateTenant(ctx, pgdb.CreateTenantParams{Slug: "refresh-test-org", Name: "Refresh Test Org"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pgdb.AdminStore().Pool.Exec(context.Background(), "DELETE FROM public.tenant WHERE id=$1", tenant.ID)
	})

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	owner, err := pgdb.AdminStore().CreateUserOwner(ctx, pgdb.CreateUserOwnerParams{
		Email:    email,
		Password: string(passwordHash),
		Name:     "testing",
		TenantID: tenant.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	r.POST("/refresh", new(AuthController).Refresh)

	// seedRefresh inserts the redis value and signs a refresh token claiming
	// the given owner.
	seedRefresh := func(keyValue string, claimed uuid.UUID, suspended bool) (string, string) {
		id := uuid.NewString()
		if err := db.GetRedis().Set(id, keyValue, time.Minute).Err(); err != nil {
			t.Fatal(err)
		}
		if suspended {
			if _, err := pgdb.AdminStore().Pool.Exec(ctx, "UPDATE public.tenant SET status=0 WHERE id=$1", tenant.ID); err != nil {
				t.Fatal(err)
			}
		}
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"refresh_uuid": id, "user_id": claimed.String(), "tenant_id": tenant.ID.String(), "exp": time.Now().Add(time.Minute).Unix(),
		}).SignedString([]byte(os.Getenv("REFRESH_SECRET")))
		if err != nil {
			t.Fatal(err)
		}
		return id, token
	}
	refresh := func(token string) *httptest.ResponseRecorder {
		body, err := json.Marshal(map[string]string{"refresh_token": token})
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest("POST", "/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	for _, tc := range []struct {
		name       string
		keyValue   string
		claimed    uuid.UUID
		wantStatus int
	}{
		{"owner mismatch", owner.String() + ":" + tenant.ID.String() + ":owner", uuid.New(), http.StatusUnauthorized},
		{"suspended tenant", owner.String() + ":" + tenant.ID.String() + ":owner", owner, http.StatusUnauthorized},
		{"matching owner", owner.String() + ":" + tenant.ID.String() + ":owner", owner, http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantStatus == http.StatusOK {
				// Only the successful case expects an active tenant.
				if _, err := pgdb.AdminStore().Pool.Exec(ctx, "UPDATE public.tenant SET status=1 WHERE id=$1", tenant.ID); err != nil {
					t.Fatal(err)
				}
			}
			suspended := tc.name == "suspended tenant"
			id, token := seedRefresh(tc.keyValue, tc.claimed, suspended)
			t.Cleanup(func() { db.GetRedis().Del(id) })

			w := refresh(token)
			if w.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tc.wantStatus)
			}
			remaining, err := db.GetRedis().Exists(id).Result()
			if err != nil {
				t.Fatal(err)
			}
			wantRemaining := int64(0)
			if tc.name == "owner mismatch" {
				wantRemaining = 1
			}
			if remaining != wantRemaining {
				t.Errorf("remaining refresh key = %d, want %d", remaining, wantRemaining)
			}
			if w.Code == http.StatusOK {
				var result struct {
					AccessToken  string `json:"access_token"`
					RefreshToken string `json:"refresh_token"`
				}
				if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
					t.Fatal(err)
				}
				accessJWT, err := jwt.Parse(result.AccessToken, func(_ *jwt.Token) (interface{}, error) {
					return []byte(os.Getenv("ACCESS_SECRET")), nil
				})
				if err != nil {
					t.Fatal(err)
				}
				claims := accessJWT.Claims.(jwt.MapClaims)
				db.GetRedis().Del(claims["access_uuid"].(string))
				if claims["user_id"] != owner.String() || claims["tenant_id"] != tenant.ID.String() || claims["role"] != "owner" {
					t.Errorf("access claims = %v, want fresh owner identity", claims)
				}
				refreshJWT, err := jwt.Parse(result.RefreshToken, func(_ *jwt.Token) (interface{}, error) {
					return []byte(os.Getenv("REFRESH_SECRET")), nil
				})
				if err != nil {
					t.Fatal(err)
				}
				refreshClaims := refreshJWT.Claims.(jwt.MapClaims)
				db.GetRedis().Del(refreshClaims["refresh_uuid"].(string))
				if replay := refresh(token); replay.Code != http.StatusUnauthorized {
					t.Errorf("replay status = %d, want 401", replay.Code)
				}
			}
			// Restore an active tenant for the remaining cases.
			if _, err := pgdb.AdminStore().Pool.Exec(ctx, "UPDATE public.tenant SET status=1 WHERE id=$1", tenant.ID); err != nil {
				t.Fatal(err)
			}
		})
	}
}
