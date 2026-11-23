package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestValidateTarget proves the guard rails of the tool: the application
// database, the protected cluster databases and any name outside the
// allowlist are refused.
func TestValidateTarget(t *testing.T) {
	cases := []struct {
		name    string
		target  string
		appDB   string
		wantErr bool
	}{
		{name: "allowed name", target: "artisan_e2e", appDB: "golang_gin_db"},
		{name: "empty name", target: "", appDB: "golang_gin_db", wantErr: true},
		{name: "blank name", target: "   ", appDB: "golang_gin_db", wantErr: true},
		{name: "application database", target: "golang_gin_db", appDB: "golang_gin_db", wantErr: true},
		{name: "application database different case", target: "GOLANG_GIN_DB", appDB: "golang_gin_db", wantErr: true},
		{name: "uppercase target", target: "Artisan_E2E", appDB: "golang_gin_db", wantErr: true},
		{name: "dash in the name", target: "artisan-e2e", appDB: "golang_gin_db", wantErr: true},
		{name: "quote in the name", target: `artisan"e2e`, appDB: "golang_gin_db", wantErr: true},
		{name: "too short", target: "ab", appDB: "golang_gin_db", wantErr: true},
		{name: "leading digit", target: "1artisan", appDB: "golang_gin_db", wantErr: true},
		{name: "maintenance database", target: "postgres", appDB: "golang_gin_db", wantErr: true},
		{name: "template database", target: "template0", appDB: "golang_gin_db", wantErr: true},
		{name: "no application database configured", target: "artisan_e2e", appDB: ""},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateTarget(testCase.target, testCase.appDB)
			if testCase.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

// TestSuperuserCredentials proves the superuser login comes from the
// explicit E2E_DB_ADMIN_* variables and falls back to the application
// credentials when they are absent.
func TestSuperuserCredentials(t *testing.T) {
	env := map[string]string{
		"DB_USER": "app_user",
		"DB_PASS": "app_user_dev_password",
	}
	getenv := func(name string) string { return env[name] }

	user, pass := superuserCredentials(getenv)
	require.Equal(t, "app_user", user)
	require.Equal(t, "app_user_dev_password", pass)

	env["E2E_DB_ADMIN_USER"] = "postgres"
	env["E2E_DB_ADMIN_PASS"] = "postgres"
	user, pass = superuserCredentials(getenv)
	require.Equal(t, "postgres", user)
	require.Equal(t, "postgres", pass)
}

// TestDSNQuotesTheValues proves the connection string quotes every value, so
// a password holding a quote or a space cannot break the DSN.
func TestDSNQuotesTheValues(t *testing.T) {
	env := map[string]string{
		"DB_HOST": "127.0.0.1",
		"DB_PORT": "5432",
		"DB_USER": "postgres",
		"DB_PASS": "p'a ss",
		"SSL":     "TRUE",
	}
	getenv := func(name string) string { return env[name] }

	got := dsn(getenv, maintenanceDB)
	want := "host='127.0.0.1' port='5432' user='postgres' password='p\\'a ss' dbname='postgres' sslmode=require"
	require.Equal(t, want, got)

	env["SSL"] = "FALSE"
	require.True(t, strings.Contains(dsn(getenv, maintenanceDB), "sslmode=disable"))
}
