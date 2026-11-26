// Command e2edb prepares a disposable environment: it recreates the target
// PostgreSQL database and flushes the Redis index named by REDIS_DB (the
// target suite's sessions index; refuses the dev index 1 when REDIS_DB is
// unset). It serves both disposable suites: the e2e cycle (artisan_e2e,
// REDIS_DB=3) and the Go suites (artisan_test, REDIS_DB=2) — each Makefile
// target pins its own pair.
//
// This binary owns DDL that the application must never run. Creating and
// dropping a database is an operations task — the same family as the goose
// migrations and the superadmin bootstrap — and not tenant business SQL: the
// tenant data only ever lives in the sqlc queries of db/queries/, and this
// tool issues no business query at all. It only manages the lifecycle of the
// test environment.
//
// The guard rails come first: the tool refuses to touch a database whose name
// is empty, whose name is the application database (DB_NAME), whose name is a
// protected cluster database or whose name does not match the strict
// allowlist ^[a-z][a-z0-9_]{2,62}$. The identifier is then quoted by
// pgx.Identifier.Sanitize, so no user-supplied value ever reaches a statement
// as raw text.
//
// Usage:
//
//	E2E_DB_NAME=artisan_e2e REDIS_DB=3 go run ./cmd/e2edb
//	E2E_DB_NAME=artisan_test REDIS_DB=2 go run ./cmd/e2edb
//
// The connection parameters are read from the environment (DB_HOST, DB_PORT,
// E2E_DB_ADMIN_USER, E2E_DB_ADMIN_PASS, DB_USER, DB_PASS, SSL, REDIS_HOST,
// REDIS_PASSWORD, REDIS_DB): the Makefile target fills them from .env and
// pins the migration superuser (postgres/postgres) and REDIS_DB to the
// target suite's index (e2e: 3, Go suites: 2). The
// superuser login comes from E2E_DB_ADMIN_* when it is set, from
// DB_USER/DB_PASS otherwise.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Massad/gin-boilerplate/internal/platform/redisdb"
	"github.com/go-redis/redis/v7"
	"github.com/jackc/pgx/v5"
)

const (
	// targetDBPattern is the strict allowlist of the target database name:
	// lowercase, no dash, no quote, no space.
	targetDBPattern = `^[a-z][a-z0-9_]{2,62}$`

	// maintenanceDB is the database every statement runs against: a database
	// cannot be dropped from inside itself.
	maintenanceDB = "postgres"

	// timeout bounds the whole preparation.
	timeout = 2 * time.Minute
)

// targetNamePattern is the compiled allowlist of the target database name.
var targetNamePattern = regexp.MustCompile(targetDBPattern)

// protectedDBNames are the cluster databases the tool must never drop, the
// maintenance database included: dropping it would break every other
// connection of the cluster.
var protectedDBNames = map[string]bool{
	maintenanceDB: true,
	"template0":   true,
	"template1":   true,
}

func main() {
	if err := run(context.Background(), os.Getenv); err != nil {
		fmt.Fprintf(os.Stderr, "e2edb: %v\n", err)
		os.Exit(1)
	}
}

// run prepares the environment: the target database is recreated and the
// Redis index named by REDIS_DB is flushed. The index is resolved before any
// destructive step, so an unset REDIS_DB is refused here.
func run(ctx context.Context, getenv func(string) string) error {
	target := strings.TrimSpace(getenv("E2E_DB_NAME"))
	if err := validateTarget(target, getenv("DB_NAME")); err != nil {
		return err
	}
	index, err := redisFlushIndex(getenv)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := recreateDatabase(ctx, getenv, target); err != nil {
		return err
	}
	if err := flushRedis(ctx, getenv, index); err != nil {
		return err
	}

	fmt.Printf("The e2e environment %s is ready\n", target)
	return nil
}

// validateTarget enforces the guard rails: an empty name, the application
// database itself, a protected cluster database, or a name outside the
// allowlist is refused before any connection is opened.
func validateTarget(target, appDB string) error {
	if target == "" {
		return errors.New("E2E_DB_NAME is empty: refusing to recreate a database")
	}
	if appDB != "" && strings.EqualFold(target, strings.TrimSpace(appDB)) {
		return fmt.Errorf("E2E_DB_NAME %q is the application database: refusing to drop it", target)
	}
	if protectedDBNames[target] {
		return fmt.Errorf("E2E_DB_NAME %q is a protected cluster database: refusing to drop it", target)
	}
	if !targetNamePattern.MatchString(target) {
		return fmt.Errorf("E2E_DB_NAME %q does not match the allowlist %s", target, targetDBPattern)
	}
	return nil
}

// redisFlushIndex resolves the Redis index through the shared parser of the
// db package and refuses the dev index when REDIS_DB is unset: the e2e cycle
// must never flush the dev sessions by accident.
func redisFlushIndex(getenv func(string) string) (int, error) {
	index, err := redisdb.RedisDBIndex(getenv)
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(getenv("REDIS_DB")) == "" && index == redisdb.DefaultRedisDB {
		return 0, errors.New("REDIS_DB is unset: refusing to flush the dev Redis index (1)")
	}
	return index, nil
}

// recreateDatabase drops the target database (forcing the disconnection of
// the sessions left over by a previous run) and creates it again, empty.
func recreateDatabase(ctx context.Context, getenv func(string) string, target string) error {
	conn, err := pgx.Connect(ctx, dsn(getenv, maintenanceDB))
	if err != nil {
		return fmt.Errorf("connect to the maintenance database: %w", err)
	}
	defer func() { _ = conn.Close(ctx) }()

	// The identifier is quoted by pgx: the name never reaches the statement as
	// raw text.
	identifier := pgx.Identifier{target}.Sanitize()

	fmt.Printf("Dropping the database %s\n", target)
	if _, err := conn.Exec(ctx, "DROP DATABASE IF EXISTS "+identifier+" WITH (FORCE)"); err != nil {
		return fmt.Errorf("drop the database %s: %w", target, err)
	}

	fmt.Printf("Creating the database %s\n", target)
	if _, err := conn.Exec(ctx, "CREATE DATABASE "+identifier); err != nil {
		return fmt.Errorf("create the database %s: %w", target, err)
	}
	return nil
}

// flushRedis empties the Redis index given by the caller: the session keys
// of a previous run must not survive.
func flushRedis(ctx context.Context, getenv func(string) string, index int) error {
	host := strings.TrimSpace(getenv("REDIS_HOST"))
	if host == "" {
		return errors.New("REDIS_HOST is empty: the e2e Redis index cannot be flushed")
	}
	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: getenv("REDIS_PASSWORD"),
		DB:       index,
	}).WithContext(ctx)
	defer func() { _ = client.Close() }()

	fmt.Printf("Flushing the Redis index %d\n", index)
	if err := client.FlushDB().Err(); err != nil {
		return fmt.Errorf("flush the Redis index %d: %w", index, err)
	}
	return nil
}

// superuserCredentials returns the login of the migration superuser: the
// explicit E2E_DB_ADMIN_* variables win, the application DB_USER/DB_PASS are
// the fallback. Only a superuser can drop and create a database.
func superuserCredentials(getenv func(string) string) (string, string) {
	if user := strings.TrimSpace(getenv("E2E_DB_ADMIN_USER")); user != "" {
		return user, getenv("E2E_DB_ADMIN_PASS")
	}
	return strings.TrimSpace(getenv("DB_USER")), getenv("DB_PASS")
}

// dsn builds the libpq keyword/value connection string of the given
// database. The values are quoted exactly like the application pools do
// (internal/platform/db).
func dsn(getenv func(string) string, database string) string {
	quote := func(value string) string {
		return "'" + strings.NewReplacer("\\", "\\\\", "'", "\\'").Replace(value) + "'"
	}
	sslMode := "disable"
	if getenv("SSL") == "TRUE" {
		sslMode = "require"
	}
	user, pass := superuserCredentials(getenv)
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		quote(getenv("DB_HOST")),
		quote(getenv("DB_PORT")),
		quote(user),
		quote(pass),
		quote(database),
		sslMode,
	)
}
