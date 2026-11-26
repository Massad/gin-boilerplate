//go:build all

package auth_test

import (
	"fmt"
	"os"
	"testing"
)

// TestMain pins the disposable environment for the whole test binary: the
// Redis index 2 and the artisan_test database. The clients must be created
// while both are present in the process environment (godotenv never
// overrides an already-set variable), so no run can drift back to the dev
// database.
func TestMain(m *testing.M) {
	if err := os.Setenv("REDIS_DB", "2"); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if err := os.Setenv("DB_NAME", "artisan_test"); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
