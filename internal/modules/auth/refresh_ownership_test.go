//go:build all

package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Massad/gin-boilerplate/controllers"
	"github.com/Massad/gin-boilerplate/internal/modules/admin"
	auth "github.com/Massad/gin-boilerplate/internal/modules/auth"
	"github.com/Massad/gin-boilerplate/internal/modules/tenant"
	user "github.com/Massad/gin-boilerplate/internal/modules/user"
	pgdb "github.com/Massad/gin-boilerplate/internal/platform/db"
	"github.com/Massad/gin-boilerplate/internal/platform/redisdb"
	testutil "github.com/Massad/gin-boilerplate/internal/testutil"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestRefreshOnlyConsumesMatchingOwner exercises the soft-consume Lua script
// and the platform pool re-verification of the refresh flow: ownership check
// on the Redis value, fresh role from the database, suspended tenant
// rejection. With the soft-consume (decision 7) the record is never deleted by
// a consumption: it stays until its TTL — reduced to the grace at the
// consumption — evicts it, so every case leaves the record in place.
func TestRefreshOnlyConsumesMatchingOwner(t *testing.T) {
	if os.Getenv("REDIS_HOST") == "" || os.Getenv("DB_HOST") == "" {
		t.Skip("set REDIS_HOST and DB_HOST to isolated services")
	}
	t.Setenv("ACCESS_SECRET", "access-key-for-isolated-tests-only-32-bytes")
	t.Setenv("REFRESH_SECRET", "refresh-key-for-isolated-tests-only-32-bytes")
	t.Setenv("ADMIN_ACCESS_SECRET", "admin-key-for-isolated-tests-only-32-byt")
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
	svc := auth.NewService(
		auth.Secrets{
			Access:      os.Getenv("ACCESS_SECRET"),
			Refresh:     os.Getenv("REFRESH_SECRET"),
			AdminAccess: os.Getenv("ADMIN_ACCESS_SECRET"),
		},
		auth.NewRepository(redisdb.GetRedis()),
	)

	// Seed a tenant and its owner on the platform pool, from the seeded
	// factories (unique slug and email per test name).
	ctx := context.Background()
	testTenant := testutil.NewTenantFixture(t)
	tenantUser, tenantForm := testutil.NewUserFixture(t)
	createdTenant, err := pgdb.AdminStore().CreateTenant(ctx, pgdb.CreateTenantParams{Slug: testTenant.Slug, Name: testTenant.Name})
	require.NoError(t, err)
	t.Cleanup(func() {
		pgdb.AdminStore().Pool.Exec(context.Background(), "DELETE FROM public.tenant WHERE id=$1", createdTenant.ID)
	})

	owner, err := pgdb.AdminStore().CreateUserOwner(ctx, pgdb.CreateUserOwnerParams{
		Email:    tenantForm.Email,
		Password: tenantUser.Password,
		Name:     tenantForm.Name,
		TenantID: createdTenant.ID,
	})
	require.NoError(t, err)

	// The refresh route wires the user and admin module services exactly
	// like the composition root does: the flow re-reads the user through
	// OneWithTenant and runs the unified login superadmin fallback through
	// the admin service.
	users := user.NewService(
		user.NewRepository(pgdb.AdminStore(), pgdb.TenantStore()),
		tenant.NewRepository(pgdb.AdminStore()),
		svc,
		pgdb.AdminStore(),
	)
	admins := admin.NewService(admin.NewRepository(pgdb.AdminStore()), tenant.NewRepository(pgdb.AdminStore()), svc, pgdb.AdminStore())
	r := gin.New()
	r.POST("/refresh", controllers.NewAuthController(svc, users, admins).Refresh)

	// seedRefresh inserts the redis value and signs a refresh token claiming
	// the given owner.
	seedRefresh := func(keyValue string, claimed uuid.UUID, suspended bool) (string, string) {
		id := uuid.NewString()
		require.NoError(t, redisdb.GetRedis().Set(id, keyValue, time.Minute).Err())
		if suspended {
			_, err := pgdb.AdminStore().Pool.Exec(ctx, "UPDATE public.tenant SET status=0 WHERE id=$1", createdTenant.ID)
			require.NoError(t, err)
		}
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"refresh_uuid": id, "user_id": claimed.String(), "tenant_id": createdTenant.ID.String(), "exp": time.Now().Add(time.Minute).Unix(),
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
		{"owner mismatch", owner.String() + ":" + createdTenant.ID.String() + ":owner", uuid.New(), http.StatusUnauthorized},
		{"suspended tenant", owner.String() + ":" + createdTenant.ID.String() + ":owner", owner, http.StatusUnauthorized},
		{"matching owner", owner.String() + ":" + createdTenant.ID.String() + ":owner", owner, http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantStatus == http.StatusOK {
				// Only the successful case expects an active tenant.
				_, err := pgdb.AdminStore().Pool.Exec(ctx, "UPDATE public.tenant SET status=1 WHERE id=$1", createdTenant.ID)
				require.NoError(t, err)
			}
			suspended := tc.name == "suspended tenant"
			id, token := seedRefresh(tc.keyValue, tc.claimed, suspended)
			t.Cleanup(func() { redisdb.GetRedis().Del(id) })

			w := refresh(token)
			require.Equal(t, tc.wantStatus, w.Code, "refresh status mismatch")
			remaining, err := redisdb.GetRedis().Exists(id).Result()
			require.NoError(t, err)
			// The soft-consume never deletes the record on a consumption:
			// the owner mismatch leaves it untouched, the suspended and
			// matching flows leave a consumed record released for the
			// former, delivered for the latter, until the TTL evicts it.
			require.Equal(t, int64(1), remaining, "remaining refresh key mismatch")
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
				require.Equal(t, createdTenant.ID.String(), claims["tenant_id"], "claim tenant_id must carry the fresh owner identity")
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
			_, err = pgdb.AdminStore().Pool.Exec(ctx, "UPDATE public.tenant SET status=1 WHERE id=$1", createdTenant.ID)
			require.NoError(t, err)
		})
	}
}
