//go:build all

package testutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	pgdb "github.com/Massad/gin-boilerplate/internal/platform/db"
)

// TestEphemeralPostgresMigrationsAndFixtures exercises the disposable instance
// end to end: migrations applied in library mode, fixtures loaded through the
// non-superuser admin pool, and tenant rows invisible for the RLS role outside
// a tenant context. Skips when Docker is unreachable.
func TestEphemeralPostgresMigrationsAndFixtures(t *testing.T) {
	ephemeral := StartEphemeralPostgres(t)
	ctx := context.Background()

	require.NoError(t, pgdb.InitAdmin(ctx, ephemeral.Env()))
	t.Cleanup(pgdb.Close)

	var version int64
	if err := pgdb.AdminStore().Pool.QueryRow(ctx, "SELECT max(version_id) FROM goose_db_version").Scan(&version); err != nil {
		require.NoError(t, err)
	}
	require.GreaterOrEqual(t, version, int64(2), "the migrations must be applied")

	LoadFixtures(t)

	var adminCount int
	require.NoError(t, pgdb.AdminStore().Pool.QueryRow(ctx, "SELECT count(*) FROM admin_user").Scan(&adminCount))
	require.Equal(t, 1, adminCount, "the admin_user fixture must be loaded")

	require.NoError(t, pgdb.InitTenant(ctx, ephemeral.Env()))
	var invisibleUserRows int
	require.NoError(t, pgdb.TenantStore().Pool.QueryRow(ctx, `SELECT count(*) FROM "user"`).Scan(&invisibleUserRows))
	require.Equal(t, 0, invisibleUserRows, "RLS must keep tenant rows invisible without a tenant context")
}
