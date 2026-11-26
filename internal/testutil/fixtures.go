package testutil

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"os"
	"path/filepath"
	"strings"
	"testing"

	testfixtures "github.com/go-testfixtures/testfixtures/v3"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"

	pgdb "github.com/Massad/gin-boilerplate/internal/platform/db"
)

// fixtureFiles is dependency ordered: the loader deletes and inserts in this
// sequence, so the foreign keys stay enforced throughout the load even though
// testfixtures cannot toggle triggers on a non-owner role.
var fixtureFiles = []string{"admin_user.yml", "tenant.yml", "user.yml", "article.yml"}

// PurgeTables empties the fixture tables through the admin pool. Plain DELETE
// statements, children first: the BYPASSRLS roles carry no TRUNCATE privilege,
// and the cascades stay active exactly as production enforces them.
func PurgeTables(t *testing.T) {
	t.Helper()
	require.NotZero(t, pgdb.AdminStore().Pool, "admin pool must be initialized before purging")
	ctx := context.Background()
	for _, table := range []string{`public.article`, `public."user"`, `public.tenant`, `public.admin_user`} {
		_, err := pgdb.AdminStore().Pool.Exec(ctx, `DELETE FROM `+table)
		require.NoError(t, err, "purging table %s", table)
	}
}

// LoadFixtures purges and loads tests/fixtures through the admin pool. It is
// called once per suite, never per sub-test.
func LoadFixtures(t *testing.T) {
	t.Helper()
	require.NotZero(t, pgdb.AdminStore().Pool, "admin pool must be initialized before loading")
	PurgeTables(t)

	connector := stdlib.GetConnector(*pgdb.AdminStore().Pool.Config().ConnConfig)
	sqlDB := sql.OpenDB(triggerTogglesConnector{connector})
	defer sqlDB.Close()

	loader, err := testfixtures.New(
		testfixtures.Database(sqlDB),
		testfixtures.Dialect("postgres"),
		testfixtures.SkipResetSequences(),
		// The shared test database (golang_gin_db) is disposable but its name
		// carries no "test" token for testfixtures' safety check.
		testfixtures.DangerousSkipTestDatabaseCheck(),
		testfixtures.Files(fixturePaths(t)...),
	)
	require.NoError(t, err)
	require.NoError(t, loader.Load())
}

// triggerTogglesConnector wraps the admin pool connector for the fixture
// loader. The testfixtures PostgreSQL helper unconditionally wraps every load
// in ALTER TABLE ... DISABLE/ENABLE TRIGGER ALL batches, which only a table
// owner may run; the loader connects as the BYPASSRLS app_admin role and can
// never own the migration-created tables. Those batches are dropped, so the
// load runs with the constraints fully enforced, in fixtureFiles order.
type triggerTogglesConnector struct {
	driver.Connector
}

func (c triggerTogglesConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.Connector.Connect(ctx)
	if err != nil {
		return nil, err
	}
	if stdConn, ok := conn.(*stdlib.Conn); ok {
		return triggerTogglesConn{stdConn}, nil
	}
	return conn, nil
}

// triggerTogglesConn intercepts only the trigger toggle batches; every other
// statement is delegated to the pgx stdlib connection.
type triggerTogglesConn struct {
	*stdlib.Conn
}

func (c triggerTogglesConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if strings.Contains(query, "DISABLE TRIGGER ALL") || strings.Contains(query, "ENABLE TRIGGER ALL") {
		return driver.RowsAffected(0), nil
	}
	return c.Conn.ExecContext(ctx, query, args)
}

// fixturePaths resolves tests/fixtures against the repository root so any
// package directory can be the test working directory.
func fixturePaths(t *testing.T) []string {
	t.Helper()
	root := repoRoot(t)
	paths := make([]string, 0, len(fixtureFiles))
	for _, file := range fixtureFiles {
		paths = append(paths, filepath.Join(root, "tests", "fixtures", file))
	}
	return paths
}

// repoRoot walks up from the working directory to the directory holding
// go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for depth := 0; depth < 10; depth++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	require.FailNow(t, "go.mod not found above the working directory")
	return ""
}
