//go:build all

package main

import (
	"context"
	"os"
	"testing"

	pgdb "github.com/Massad/gin-boilerplate/internal/db"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

// TestRunCreateAdmin covers the CLI bootstrap: two identical calls succeed
// (idempotent through ON CONFLICT (email) DO UPDATE) and the stored hash is a
// valid bcrypt hash of the last password.
func TestRunCreateAdmin(t *testing.T) {
	if os.Getenv("DB_HOST") == "" || os.Getenv("DB_ADMIN_USER") == "" {
		if err := godotenv.Load(".env"); err != nil {
			t.Skip("set DB_HOST/DB_ADMIN_USER or create a .env file")
		}
	}
	if os.Getenv("DB_HOST") == "" || os.Getenv("DB_ADMIN_USER") == "" {
		t.Skip("set DB_HOST/DB_ADMIN_USER to an isolated PostgreSQL instance")
	}

	email := "bootstrap-admin@test.com"
	firstPassword := "first-bootstrap-password"
	secondPassword := "second-bootstrap-password"
	ctx := context.Background()

	if err := runCreateAdmin(ctx, os.Getenv, email, firstPassword); err != nil {
		t.Fatalf("first bootstrap failed: %v", err)
	}
	if err := runCreateAdmin(ctx, os.Getenv, email, secondPassword); err != nil {
		t.Fatalf("second bootstrap must succeed (idempotent): %v", err)
	}

	// runCreateAdmin closed its pools: reconnect to verify the stored hash.
	if err := pgdb.InitAdmin(ctx, os.Getenv); err != nil {
		t.Fatal(err)
	}
	defer pgdb.Close()
	row, err := pgdb.AdminStore().GetAdminByEmail(ctx, email)
	if err != nil {
		t.Fatal(err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(row.Password), []byte(secondPassword)); err != nil {
		t.Fatalf("stored hash does not match the last bootstrap password: %v", err)
	}
	t.Cleanup(func() {
		pgdb.AdminStore().Pool.Exec(context.Background(), "DELETE FROM public.admin_user WHERE email=$1", email)
	})

	// Missing arguments fail without touching the database.
	if err := runCreateAdmin(ctx, os.Getenv, "", secondPassword); err == nil {
		t.Fatal("empty email must be rejected")
	}
}
