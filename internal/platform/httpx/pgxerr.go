package httpx

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// MapPgxError maps a pgx error to the HTTP status the project conventions
// define: a missing row is 404, unique and foreign key violations are 409 and
// a row level security violation is 500. Any other error maps to 500; the raw
// SQL state never reaches the client.
func MapPgxError(err error) int {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505", "23503":
			return http.StatusConflict
		case "42501":
			return http.StatusInternalServerError
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}
