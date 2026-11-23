package main

import (
	"strings"
	"testing"
)

// TestValidateTarget proves the guard rails of the tool: the application
// database and any name outside the allowlist are refused.
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
		{name: "no application database configured", target: "artisan_e2e", appDB: ""},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateTarget(testCase.target, testCase.appDB)
			if testCase.wantErr && err == nil {
				t.Fatalf("validateTarget(%q, %q) = nil, want an error", testCase.target, testCase.appDB)
			}
			if !testCase.wantErr && err != nil {
				t.Fatalf("validateTarget(%q, %q) = %v, want nil", testCase.target, testCase.appDB, err)
			}
		})
	}
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
	if got != want {
		t.Fatalf("dsn() = %q, want %q", got, want)
	}

	env["SSL"] = "FALSE"
	if !strings.Contains(dsn(getenv, maintenanceDB), "sslmode=disable") {
		t.Fatal("dsn() must disable the TLS when SSL is not TRUE")
	}
}
