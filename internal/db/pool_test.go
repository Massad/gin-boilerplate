//go:build all

package db

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
		t.Fatal(err)
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
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close(ctx)
			var name string
			if err := conn.QueryRow(ctx, "SELECT current_database()").Scan(&name); err != nil {
				t.Fatal(err)
			}
			if name != os.Getenv("DB_NAME") {
				t.Fatalf("connected to %q, want configured database", name)
			}
		})
	}
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
