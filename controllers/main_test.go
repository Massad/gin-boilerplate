//go:build all

package controllers

import (
	"fmt"
	"os"
	"testing"
)

// TestMain pins the disposable Redis index for the whole test binary: the
// client must be created while REDIS_DB is present in the process
// environment (godotenv never overrides an already-set variable).
func TestMain(m *testing.M) {
	if err := os.Setenv("REDIS_DB", "2"); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
