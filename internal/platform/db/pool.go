package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store bundles the sqlc queries with the underlying pgx pool.
type Store struct {
	*Queries
	Pool *pgxpool.Pool
	Role string
}

// Connect opens a pgx pool for the given DSN and pings it before returning.
func Connect(ctx context.Context, dsn string) (Store, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return Store{}, err
	}
	cfg.MaxConns = 10
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return Store{}, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return Store{}, err
	}
	return Store{Queries: New(pool), Pool: pool}, nil
}

// InTx runs fn inside a transaction, handing it the raw transaction and the
// queries bound to it.
func (s Store) InTx(ctx context.Context, fn func(tx pgx.Tx, q *Queries) error) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := fn(tx, s.Queries.WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// InTx is the narrow view of an ongoing transaction handed to the domain
// code: the tx-bound queries are only reached inside the code of the module
// executing them — never a parameter or return value of a cross-module port
// (design decision 6).
type InTx interface {
	Queries() *Queries
}

// withinTx is the InTx view of one running transaction: it carries nothing
// but the queries bound to it, the Begin/Rollback/Commit mechanics stay
// inside WithinTx.
type withinTx struct {
	queries *Queries
}

// Queries returns the queries bound to the running transaction.
func (w withinTx) Queries() *Queries { return w.queries }

// WithinTx runs fn inside a transaction, handing it the narrow InTx view.
// Strict wrapper of Store.InTx: the Begin/Rollback/Commit mechanics are
// identical and no set_config is issued — the in-Tx register operations
// target the admin BYPASSRLS pool without tenant context.
func (s Store) WithinTx(ctx context.Context, fn func(tx InTx) error) error {
	return s.InTx(ctx, func(tx pgx.Tx, q *Queries) error {
		return fn(withinTx{queries: q})
	})
}

// WithTenant runs fn inside a transaction with the RLS context set for
// tenantID. The context is configured with set_config(..., true) so it resets
// automatically on commit; a session-wide SET is forbidden on a shared pool.
func (s Store) WithTenant(ctx context.Context, tenantID uuid.UUID, fn func(q *Queries) error) error {
	return s.InTx(ctx, func(tx pgx.Tx, q *Queries) error {
		if _, err := tx.Exec(ctx,
			`SELECT set_config('app.tenant_id', $1, true)`,
			tenantID.String()); err != nil {
			return err
		}
		return fn(q)
	})
}

// postgresDSN builds a libpq keyword/value DSN from environment variables. The
// pgx parser follows the same rules as the previous lib/pq driver, so the
// quoting of every value (including empty strings) is preserved.
func postgresDSN(getenv func(string) string, user, pass string) string {
	sslMode := "disable"
	if getenv("SSL") == "TRUE" {
		sslMode = "require"
	}

	quote := func(value string) string {
		return "'" + strings.NewReplacer("\\", "\\\\", "'", "\\'").Replace(value) + "'"
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		quote(getenv("DB_HOST")),
		quote(getenv("DB_PORT")),
		quote(user),
		quote(pass),
		quote(getenv("DB_NAME")),
		sslMode,
	)
}

var tenantStore, adminStore Store

// InitTenant connects the tenant pool (application role, RLS constrained).
func InitTenant(ctx context.Context, getenv func(string) string) error {
	s, err := Connect(ctx, postgresDSN(getenv, getenv("DB_USER"), getenv("DB_PASS")))
	if err != nil {
		return err
	}
	s.Role = "tenant"
	tenantStore = s
	return nil
}

// InitAdmin connects the platform/admin pool (BYPASSRLS role).
func InitAdmin(ctx context.Context, getenv func(string) string) error {
	s, err := Connect(ctx, postgresDSN(getenv, getenv("DB_ADMIN_USER"), getenv("DB_ADMIN_PASS")))
	if err != nil {
		return err
	}
	s.Role = "admin"
	adminStore = s
	return nil
}

// TenantStore returns the tenant pool store. The generated models declare a
// Tenant struct, so the accessor carries the Store suffix.
func TenantStore() Store { return tenantStore }

// AdminStore returns the platform/admin pool store.
func AdminStore() Store { return adminStore }

// Close releases both pools.
func Close() {
	if tenantStore.Pool != nil {
		tenantStore.Pool.Close()
	}
	if adminStore.Pool != nil {
		adminStore.Pool.Close()
	}
}
