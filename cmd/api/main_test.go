//go:build all

package main

import (
	"context"
	"os"
	"testing"

	pgdb "github.com/Massad/gin-boilerplate/internal/platform/db"
	testutil "github.com/Massad/gin-boilerplate/internal/testutil"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// TestRunCreateAdmin covers the CLI bootstrap: two identical calls succeed
// (idempotent through ON CONFLICT (email) DO UPDATE) and the stored hash is a
// valid bcrypt hash of the last password. The -name flag is stored through the
// same upsert (the last name wins) and an empty -name falls back to the
// default « Superadmin ».
func TestRunCreateAdmin(t *testing.T) {
	if os.Getenv("DB_HOST") == "" || os.Getenv("DB_ADMIN_USER") == "" {
		if err := godotenv.Load(".env"); err != nil {
			t.Skip("set DB_HOST/DB_ADMIN_USER or create a .env file")
		}
	}
	if os.Getenv("DB_HOST") == "" || os.Getenv("DB_ADMIN_USER") == "" {
		t.Skip("set DB_HOST/DB_ADMIN_USER to an isolated PostgreSQL instance")
	}

	email, password := testutil.NewAdminFixture(t)
	// The CLI idempotently rotates the stored hash: the last password wins.
	firstPassword := password
	secondPassword := password + "-rotated"
	firstName := "Prénom Nom"
	secondName := "Laure Moreau"
	ctx := context.Background()

	require.NoError(t, runCreateAdmin(ctx, os.Getenv, email, firstPassword, firstName), "first bootstrap failed")
	require.NoError(t, runCreateAdmin(ctx, os.Getenv, email, secondPassword, secondName), "second bootstrap must succeed (idempotent)")

	// runCreateAdmin closed its pools: reconnect to verify the stored identity.
	require.NoError(t, pgdb.InitAdmin(ctx, os.Getenv))
	row, err := pgdb.AdminStore().GetAdminByEmail(ctx, email)
	require.NoError(t, err)
	require.Equal(t, secondName, row.Name, "the last bootstrap name must win (upsert with -name)")
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(row.Password), []byte(secondPassword)), "stored hash must match the last bootstrap password")
	pgdb.Close()

	// An empty -name falls back to the default display name.
	require.NoError(t, runCreateAdmin(ctx, os.Getenv, email, secondPassword, ""), "empty -name accepted with the default")
	require.NoError(t, pgdb.InitAdmin(ctx, os.Getenv))
	defer pgdb.Close()
	row, err = pgdb.AdminStore().GetAdminByEmail(ctx, email)
	require.NoError(t, err)
	require.Equal(t, "Superadmin", row.Name, "the default display name is stored when -name is empty")
	t.Cleanup(func() {
		pgdb.AdminStore().Pool.Exec(context.Background(), "DELETE FROM public.admin_user WHERE email=$1", email)
	})

	// Missing arguments fail without touching the database.
	require.Error(t, runCreateAdmin(ctx, os.Getenv, "", secondPassword, secondName), "empty email must be rejected")
}
