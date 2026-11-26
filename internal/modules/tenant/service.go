// Package tenant holds the tenant domain: slug derivation rules, in-Tx
// provisioning consumed by the register flow and the platform backoffice
// data. The tenant table is a platform table: every operation targets the
// admin pool, never a tenant context.
package tenant

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	pgdb "github.com/Massad/gin-boilerplate/internal/platform/db"
	"github.com/jackc/pgx/v5/pgconn"

	"golang.org/x/text/unicode/norm"
)

// ErrInvalidSlug is returned when a company name does not map to a usable slug.
var ErrInvalidSlug = errors.New("the company name does not map to a valid slug")

// ErrSlugUnavailable is returned when every variation of the slug is taken.
var ErrSlugUnavailable = errors.New("a slug could not be generated for this company name")

// reservedSlugs lists route roots already used by the platform: a tenant path
// must never collide with them. The live route roots are the first segments
// of the frozen route list pinned by the API contract test TestAPISurface
// (tests/api_surface_test.go, P0.3): when a new root shows up there, it must
// be mirrored here, otherwise a tenant could register its slug and shadow the
// route.
var reservedSlugs = map[string]bool{
	"v1":      true,
	"admin":   true,
	"public":  true,
	"swagger": true,
	"api":     true,
	"login":   true,
	"app":     true,
	"sw":      true,
	"static":  true,
	"www":     true,
	"user":    true,
	"article": true,
	"invoice": true,
	"token":   true,
}

// tenantSlugPattern bounds a slug between 3 and 40 characters, letters, digits
// and dashes only, without leading or trailing dash.
var tenantSlugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,38}[a-z0-9]$`)

// maxSlugLength is the longest slug accepted by tenantSlugPattern.
const maxSlugLength = 40

// maxSlugAttempts bounds the -2, -3, ... suffix search before giving up.
const maxSlugAttempts = 99

// Slugify turns a company name into its normalized slug: lowercase, without
// accents, with any other character run collapsed into a single dash.
func Slugify(name string) string {
	var b strings.Builder
	for _, r := range norm.NFD.String(strings.ToLower(name)) {
		switch {
		case unicode.Is(unicode.Mn, r): // combining mark: part of an accent
			continue
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	slug := dashRuns.ReplaceAllString(b.String(), "-")
	return strings.Trim(slug, "-")
}

// dashRuns collapses consecutive separator characters into a single dash.
var dashRuns = regexp.MustCompile("-{2,}")

// ValidSlug reports whether the slug is well formed and not reserved.
func ValidSlug(slug string) bool {
	return tenantSlugPattern.MatchString(slug) && !reservedSlugs[slug]
}

// shortSlugCandidate builds suffixed candidates that keep the slug within the
// maximum length accepted by the pattern.
func shortSlugCandidate(base string, attempt int) string {
	suffix := "-" + strconv.Itoa(attempt)
	if len(base)+len(suffix) > maxSlugLength {
		base = strings.Trim(base[:maxSlugLength-len(suffix)], "-")
	}
	return base + suffix
}

// createTenantInTx inserts the tenant inside the ongoing platform transaction,
// retrying on slug collisions with the standard candidates. The tx-bound
// *pgdb.Queries is reached only here, inside the executing module code.
func createTenantInTx(ctx context.Context, q *pgdb.Queries, name string) (Tenant, error) {
	base := Slugify(name)
	if !ValidSlug(base) {
		return Tenant{}, ErrInvalidSlug
	}
	for attempt := 1; attempt <= maxSlugAttempts; attempt++ {
		candidate := base
		if attempt > 1 {
			candidate = shortSlugCandidate(base, attempt)
		}
		row, err := q.CreateTenant(ctx, pgdb.CreateTenantParams{Slug: candidate, Name: name})
		if err == nil {
			return newTenant(row.ID, row.Slug, row.Name, row.Status, row.CreatedAt, row.UpdatedAt), nil
		}
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
			return Tenant{}, err
		}
	}
	return Tenant{}, ErrSlugUnavailable
}
