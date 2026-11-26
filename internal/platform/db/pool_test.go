//go:build all

package db

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// Run against an isolated PostgreSQL instance with trust authentication so the
// password variations exercise the driver's DSN parser rather than credentials.
func TestPostgresDSNPasswords(t *testing.T) {
	if os.Getenv("DB_HOST") == "" || os.Getenv("DB_NAME") == "" {
		t.Skip("configure an isolated PostgreSQL instance")
	}
	// Control connection with the configured credentials: a password-auth
	// failure here means the instance does not use trust authentication and
	// the password variations below would fail for the wrong reason.
	ctx := context.Background()
	control, err := pgx.Connect(ctx, postgresDSN(os.Getenv, os.Getenv("DB_USER"), os.Getenv("DB_PASS")))
	if err != nil {
		var pgErr *pgconn.PgError
		if errorsAsPgx(err, &pgErr) && (pgErr.Code == "28P01" || pgErr.Code == "28000") {
			t.Skipf("PostgreSQL reachable but authentication is %s: configure a trust-auth instance", pgErr.Code)
		}
		require.NoError(t, err)
	}
	control.Close(ctx)

	for _, password := range []string{"", "test with spaces", "' dbname=postgres", "\\' quoted password"} {
		t.Run(password, func(t *testing.T) {
			dsn := postgresDSN(func(name string) string {
				if name == "DB_PASS" {
					return password
				}
				return os.Getenv(name)
			}, os.Getenv("DB_USER"), password)
			conn, err := pgx.Connect(ctx, dsn)
			require.NoError(t, err)
			defer conn.Close(ctx)
			var name string
			require.NoError(t, conn.QueryRow(ctx, "SELECT current_database()").Scan(&name))
			require.Equal(t, os.Getenv("DB_NAME"), name, "connected to %q, want configured database", name)
		})
	}
}

// errWithinTxProbe is returned by the rollback test: WithinTx must propagate
// the callback failure and discard the transaction.
var errWithinTxProbe = errors.New("within tx rollback probe")

// withinTxTestStore connects an admin-credential store for the WithinTx tests.
// The rows the tests insert all carry the pooltest-within-tx- prefix and the
// cleanup removes them again: the configured base keeps no residue beyond the
// scope of a test run.
func withinTxTestStore(t *testing.T) Store {
	t.Helper()
	if os.Getenv("DB_HOST") == "" || os.Getenv("DB_NAME") == "" || os.Getenv("DB_ADMIN_USER") == "" {
		t.Skip("configure a migrated PostgreSQL instance with admin credentials")
	}
	ctx := context.Background()
	store, err := Connect(ctx, postgresDSN(os.Getenv, os.Getenv("DB_ADMIN_USER"), os.Getenv("DB_ADMIN_PASS")))
	if err != nil {
		var pgErr *pgconn.PgError
		if errorsAsPgx(err, &pgErr) && (pgErr.Code == "28P01" || pgErr.Code == "28000") {
			t.Skipf("PostgreSQL reachable but authentication is %s: configure reachable credentials", pgErr.Code)
		}
		require.NoError(t, err)
	}
	t.Cleanup(func() {
		_, derr := store.Pool.Exec(ctx, `DELETE FROM tenant WHERE slug LIKE 'pooltest-within-tx-%'`)
		require.NoError(t, derr)
		store.Pool.Close()
	})
	return store
}

// withinTxSlug returns a unique tenant slug carrying the cleanup prefix.
func withinTxSlug() string {
	return "pooltest-within-tx-" + strings.ReplaceAll(uuid.NewString(), "-", "")
}

// TestWithinTxCommitsAfterFnSuccess proves the commit path: fn runs without
// error, the insert survives the transaction and stays readable on the pool.
func TestWithinTxCommitsAfterFnSuccess(t *testing.T) {
	store := withinTxTestStore(t)
	ctx := context.Background()
	slug := withinTxSlug()

	var txID uuid.UUID
	require.NoError(t, store.WithinTx(ctx, func(tx InTx) error {
		row, err := tx.Queries().CreateTenant(ctx, CreateTenantParams{Slug: slug, Name: "pool test commit"})
		if err != nil {
			return err
		}
		txID = row.ID
		return nil
	}))

	row, err := store.GetTenantByID(ctx, txID)
	require.NoError(t, err)
	require.Equal(t, slug, row.Slug)
}

// TestWithinTxRollsBackAfterFnError proves the rollback path: the fn error
// propagates and the insert made before it is discarded — the exact same slug
// becomes creatable again outside any transaction (UNIQUE(slug) intact).
func TestWithinTxRollsBackAfterFnError(t *testing.T) {
	store := withinTxTestStore(t)
	ctx := context.Background()
	slug := withinTxSlug()

	err := store.WithinTx(ctx, func(tx InTx) error {
		if _, err := tx.Queries().CreateTenant(ctx, CreateTenantParams{Slug: slug, Name: "pool test rollback"}); err != nil {
			return err
		}
		return errWithinTxProbe
	})
	require.ErrorIs(t, err, errWithinTxProbe)

	row, err := store.CreateTenant(ctx, CreateTenantParams{Slug: slug, Name: "pool test rollback after"})
	require.NoError(t, err)
	require.Equal(t, slug, row.Slug)
}

// TestWithinTxNeverSetsSessionWideTenantSetting pins withinTx to the admin
// pool without tenant context: no session-wide app.tenant_id may survive on
// the released connection. A single-connection pool makes every probe hit the
// same physical connection the transactions ran on.
func TestWithinTxNeverSetsSessionWideTenantSetting(t *testing.T) {
	if os.Getenv("DB_HOST") == "" || os.Getenv("DB_NAME") == "" || os.Getenv("DB_ADMIN_USER") == "" {
		t.Skip("configure a migrated PostgreSQL instance with admin credentials")
	}
	ctx := context.Background()
	cfg, err := pgxpool.ParseConfig(postgresDSN(os.Getenv, os.Getenv("DB_ADMIN_USER"), os.Getenv("DB_ADMIN_PASS")))
	require.NoError(t, err)
	cfg.MaxConns = 1
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errorsAsPgx(err, &pgErr) && (pgErr.Code == "28P01" || pgErr.Code == "28000") {
			t.Skipf("PostgreSQL reachable but authentication is %s: configure reachable credentials", pgErr.Code)
		}
		require.NoError(t, err)
	}
	t.Cleanup(pool.Close)

	// sessionSetting reads the surviving app.tenant_id value of the pooled
	// connection: empty means no session-wide state leaked.
	sessionSetting := func() string {
		var value string
		require.NoError(t, pool.QueryRow(ctx, `SELECT coalesce(current_setting('app.tenant_id', true), '')`).Scan(&value))
		return value
	}

	// Probe sensitivity control: a real session-wide write must be visible,
	// then be cleared again.
	sensing := withinTxSlug()
	_, err = pool.Exec(ctx, "SELECT set_config('app.tenant_id', $1, false)", sensing)
	require.NoError(t, err)
	require.Equal(t, sensing, sessionSetting(), "the probe must see a session-wide app.tenant_id")
	_, err = pool.Exec(ctx, "SELECT set_config('app.tenant_id', '', false)")
	require.NoError(t, err)

	var store = Store{Queries: New(pool), Pool: pool}

	// WithinTx: the transaction commits without any surviving tenant context.
	slug := withinTxSlug()
	require.NoError(t, store.WithinTx(ctx, func(tx InTx) error {
		_, err := tx.Queries().CreateTenant(ctx, CreateTenantParams{Slug: slug, Name: "pool test tenant setting"})
		return err
	}))
	require.Equal(t, "", sessionSetting(), "WithinTx leaked a session-wide app.tenant_id")

	// Control: WithTenant sets the value locally inside its transaction and
	// the local set_config reverts on commit — nothing survives a release.
	require.NoError(t, store.WithTenant(ctx, uuid.New(), func(q *Queries) error {
		_, lerr := q.ListTenants(ctx)
		return lerr
	}))
	require.Equal(t, "", sessionSetting(), "WithTenant leaked a session-wide app.tenant_id")
}

func errorsAsPgx(err error, target **pgconn.PgError) bool {
	if e, ok := err.(*pgconn.PgError); ok {
		*target = e
		return true
	}
	if err != nil {
		// Unwrap one level for wrapped pgxv5 errors.
		if u, ok := err.(interface{ Unwrap() error }); ok {
			return errorsAsPgx(u.Unwrap(), target)
		}
	}
	return false
}
