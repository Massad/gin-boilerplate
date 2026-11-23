package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ephemeralPostgresImage is the frozen disposable PostgreSQL used opt-in.
const ephemeralPostgresImage = "postgres:16-alpine"

// EphemeralPostgres carries the connection details of the disposable
// instance: the raw host/port consumed by Env to wire the migration-created
// application roles, and the migration superuser's own DSN helper.
type EphemeralPostgres struct {
	Host string
	Port string
}

// StartEphemeralPostgres boots a disposable PostgreSQL and applies the project
// migrations through goose in library mode (the same pinned version as the
// Makefile CLI), then returns its connection details. testcontainers exposes
// no daemon probe: a startup failure means Docker is unreachable and the
// calling test skips.
func StartEphemeralPostgres(t *testing.T) EphemeralPostgres {
	t.Helper()
	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        ephemeralPostgresImage,
			Env:          map[string]string{"POSTGRES_USER": "goose", "POSTGRES_PASSWORD": "ephemeral-test-password", "POSTGRES_DB": "ephemeral_gin_boilerplate"},
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForListeningPort("5432/tcp"),
		},
		Started: true,
	})
	if err != nil {
		// Startup is the first point where the Docker daemon is actually probed.
		t.Skipf("Docker unavailable: %v", err)
		return EphemeralPostgres{}
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)
	ephemeral := EphemeralPostgres{Host: host, Port: port.Port()}

	// The migration superuser owns the tables; consumers connect as the
	// migration-created app_user/app_admin roles exactly like the shared DB.
	db, err := sql.Open("pgx", ephemeral.migrationDSN())
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, db.PingContext(ctx))

	provider, err := goose.NewProvider(goose.DialectPostgres, db, os.DirFS(filepath.Join(repoRoot(t), "db", "migrations")))
	require.NoError(t, err)
	results, err := provider.Up(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, results, "the project migrations must apply to a fresh database")
	return ephemeral
}

// Env returns a getenv func wiring pgdb.InitTenant and pgdb.InitAdmin to the
// container roles created by migration 000002.
func (e EphemeralPostgres) Env() func(string) string {
	return func(name string) string {
		switch name {
		case "DB_HOST":
			return e.Host
		case "DB_PORT":
			return e.Port
		case "DB_NAME":
			return "ephemeral_gin_boilerplate"
		case "SSL":
			return "FALSE"
		case "DB_USER":
			return "app_user"
		case "DB_PASS":
			return "app_user_dev_password"
		case "DB_ADMIN_USER":
			return "app_admin"
		case "DB_ADMIN_PASS":
			return "app_admin_dev_password"
		default:
			return ""
		}
	}
}

// migrationDSN connects as the container's migration superuser.
func (e EphemeralPostgres) migrationDSN() string {
	return fmt.Sprintf("host=%s port=%s user=goose password='ephemeral-test-password' dbname=ephemeral_gin_boilerplate sslmode=disable", e.Host, e.Port)
}
