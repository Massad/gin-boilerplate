//go:build all

package db

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
)

// Run against an isolated PostgreSQL instance with trust authentication so the
// password variations exercise the driver's DSN parser rather than credentials.
func TestPostgresDSNPasswords(t *testing.T) {
	if os.Getenv("DB_HOST") == "" || os.Getenv("DB_NAME") == "" {
		t.Skip("configure an isolated PostgreSQL instance")
	}
	for _, password := range []string{"", "test with spaces", "' dbname=postgres", "\\' quoted password"} {
		t.Run(password, func(t *testing.T) {
			t.Setenv("DB_PASS", password)
			conn, err := sqlx.Connect("postgres", postgresDSN(os.Getenv))
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			var name string
			if err := conn.Get(&name, "SELECT current_database()"); err != nil {
				t.Fatal(err)
			}
			if name != os.Getenv("DB_NAME") {
				t.Fatalf("connected to %q, want configured database", name)
			}
		})
	}
}
