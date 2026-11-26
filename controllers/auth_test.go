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

	pgdb "github.com/Massad/gin-boilerplate/internal/platform/db"
	"github.com/Massad/gin-boilerplate/internal/platform/redisdb"
	testutil "github.com/Massad/gin-boilerplate/internal/testutil"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, redisdb.InitRedis(), "resolving the Redis database index")
	require.Equal(t, 2, redisdb.GetRedis().Options().DB, "the suite must run on the disposable Redis index 2")
	t.Cleanup(func() { _ = redisdb.GetRedis().Close() })
	require.NoError(t, redisdb.GetRedis().Ping().Err())
	require.NoError(t, pgdb.InitTenant(context.Background(), os.Getenv), "initializing the tenant pool: if the error says the database does not exist, run make test once to provision artisan_test; never falls back to the dev database")
	require.NoError(t, pgdb.InitAdmin(context.Background(), os.Getenv), "initializing the admin pool: if the error says the database does not exist, run make test once to provision artisan_test; never falls back to the dev database")
	var suiteDB string
	require.NoError(t, pgdb.AdminStore().Pool.QueryRow(context.Background(), "SELECT current_database()").Scan(&suiteDB))
	require.Equal(t, "artisan_test", suiteDB, "the suite must run on the disposable database artisan_test")

	// Seed a tenant and its owner on the platform pool, from the seeded
	// factories (unique slug and email per test name).
	ctx := context.Background()
	testTenant := testutil.NewTenantFixture(t)
	tenantUser, tenantForm := testutil.NewUserFixture(t)
	tenant, err := pgdb.AdminStore().CreateTenant(ctx, pgdb.CreateTenantParams{Slug: testTenant.Slug, Name: testTenant.Name})
	require.NoError(t, err)
	t.Cleanup(func() {
		pgdb.AdminStore().Pool.Exec(context.Background(), "DELETE FROM public.tenant WHERE id=$1", tenant.ID)
	})

	owner, err := pgdb.AdminStore().CreateUserOwner(ctx, pgdb.CreateUserOwnerParams{
		Email:    tenantForm.Email,
		Password: tenantUser.Password,
		Name:     tenantForm.Name,
		TenantID: tenant.ID,
	})
	require.NoError(t, err)

	r := gin.New()
	r.POST("/refresh", new(AuthController).Refresh)

	// seedRefresh inserts the redis value and signs a refresh token claiming
	// the given owner.
	seedRefresh := func(keyValue string, claimed uuid.UUID, suspended bool) (string, string) {
		id := uuid.NewString()
		require.NoError(t, redisdb.GetRedis().Set(id, keyValue, time.Minute).Err())
		if suspended {
			_, err := pgdb.AdminStore().Pool.Exec(ctx, "UPDATE public.tenant SET status=0 WHERE id=$1", tenant.ID)
			require.NoError(t, err)
		}
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"refresh_uuid": id, "user_id": claimed.String(), "tenant_id": tenant.ID.String(), "exp": time.Now().Add(time.Minute).Unix(),
		}).SignedString([]byte(os.Getenv("REFRESH_SECRET")))
		require.NoError(t, err)
		return id, token
	}
	refresh := func(token string) *httptest.ResponseRecorder {
		body, err := json.Marshal(map[string]string{"refresh_token": token})
		require.NoError(t, err)
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
				_, err := pgdb.AdminStore().Pool.Exec(ctx, "UPDATE public.tenant SET status=1 WHERE id=$1", tenant.ID)
				require.NoError(t, err)
			}
			suspended := tc.name == "suspended tenant"
			id, token := seedRefresh(tc.keyValue, tc.claimed, suspended)
			t.Cleanup(func() { redisdb.GetRedis().Del(id) })

			w := refresh(token)
			require.Equal(t, tc.wantStatus, w.Code, "refresh status mismatch")
			remaining, err := redisdb.GetRedis().Exists(id).Result()
			require.NoError(t, err)
			wantRemaining := int64(0)
			if tc.name == "owner mismatch" {
				wantRemaining = 1
			}
			require.Equal(t, wantRemaining, remaining, "remaining refresh key mismatch")
			if w.Code == http.StatusOK {
				var result struct {
					AccessToken  string `json:"access_token"`
					RefreshToken string `json:"refresh_token"`
				}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
				accessJWT, err := jwt.Parse(result.AccessToken, func(_ *jwt.Token) (interface{}, error) {
					return []byte(os.Getenv("ACCESS_SECRET")), nil
				})
				require.NoError(t, err)
				claims := accessJWT.Claims.(jwt.MapClaims)
				redisdb.GetRedis().Del(claims["access_uuid"].(string))
				require.Equal(t, owner.String(), claims["user_id"], "claim user_id must carry the fresh owner identity")
				require.Equal(t, tenant.ID.String(), claims["tenant_id"], "claim tenant_id must carry the fresh owner identity")
				require.Equal(t, "owner", claims["role"], "claim role must carry the fresh owner identity")
				refreshJWT, err := jwt.Parse(result.RefreshToken, func(_ *jwt.Token) (interface{}, error) {
					return []byte(os.Getenv("REFRESH_SECRET")), nil
				})
				require.NoError(t, err)
				refreshClaims := refreshJWT.Claims.(jwt.MapClaims)
				redisdb.GetRedis().Del(refreshClaims["refresh_uuid"].(string))
				replay := refresh(token)
				require.Equal(t, http.StatusUnauthorized, replay.Code, "replay status must be unauthorized")
			}
			// Restore an active tenant for the remaining cases.
			_, err = pgdb.AdminStore().Pool.Exec(ctx, "UPDATE public.tenant SET status=1 WHERE id=$1", tenant.ID)
			require.NoError(t, err)
		})
	}
}
